package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"path/filepath"
	"strings"
)

const (
	// MaxImageBytes caps uploaded image size at 25 MB (within Fiber/Nginx 32 MB body limit).
	MaxImageBytes = 25 * 1024 * 1024
	// MaxImageDimension prevents decompression-bomb attacks on width or height.
	MaxImageDimension = 12000
	// MaxImagePixels caps decoded pixel count at 64 megapixels.
	MaxImagePixels = 64_000_000
)

// SanitizeAndStripMetadata validates binary image signatures, blocks script/SVG/polyglot
// payloads and decompression bombs, strips 100% of metadata (EXIF, GPS, camera serials,
// XMP, IPTC, ICC profiles, PNG text chunks, GIF comments, and trailing bytes), and
// returns clean image bytes along with the canonical file extension.
func SanitizeAndStripMetadata(raw []byte, originalFilename string) ([]byte, string, error) {
	if len(raw) == 0 {
		return nil, "", errors.New("empty image file")
	}
	if len(raw) > MaxImageBytes {
		return nil, "", fmt.Errorf("image size exceeds %dMB limit", MaxImageBytes/(1024*1024))
	}

	// Reject explicitly disallowed extensions upfront (e.g., .svg, .html, .php, .js, .exe)
	if originalFilename != "" {
		ext := strings.ToLower(filepath.Ext(filepath.Base(originalFilename)))
		if ext != "" {
			switch ext {
			case ".jpg", ".jpeg", ".jfif", ".png", ".webp", ".gif":
			default:
				return nil, "", errors.New("only JPG, JPEG, JFIF, PNG, WEBP, or GIF images are permitted")
			}
		}
	}

	// Detect true binary image format from magic bytes rather than trusting client headers/extensions
	switch {
	case len(raw) >= 3 && raw[0] == 0xFF && raw[1] == 0xD8 && raw[2] == 0xFF:
		clean, err := stripJPEGMetadata(raw)
		if err != nil {
			return nil, "", err
		}
		return clean, ".jpg", nil

	case len(raw) >= 8 && bytes.Equal(raw[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		clean, err := stripPNGMetadata(raw)
		if err != nil {
			return nil, "", err
		}
		return clean, ".png", nil

	case len(raw) >= 6 && (bytes.Equal(raw[:6], []byte("GIF87a")) || bytes.Equal(raw[:6], []byte("GIF89a"))):
		clean, err := stripGIFMetadata(raw)
		if err != nil {
			return nil, "", err
		}
		return clean, ".gif", nil

	case len(raw) >= 12 && bytes.Equal(raw[:4], []byte("RIFF")) && bytes.Equal(raw[8:12], []byte("WEBP")):
		clean, err := StripWebPMetadata(raw)
		if err != nil {
			return nil, "", err
		}
		return clean, ".webp", nil

	default:
		return nil, "", errors.New("invalid or unsupported image format")
	}
}

func validateRasterDimensions(r *bytes.Reader) error {
	cfg, _, err := image.DecodeConfig(r)
	if err != nil {
		return fmt.Errorf("corrupt or invalid image header: %w", err)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > MaxImageDimension || cfg.Height > MaxImageDimension {
		return fmt.Errorf("image dimensions (%dx%d) exceed allowed limits", cfg.Width, cfg.Height)
	}
	if int64(cfg.Width)*int64(cfg.Height) > MaxImagePixels {
		return errors.New("image pixel count exceeds safety limit")
	}
	_, _ = r.Seek(0, 0)
	return nil
}

func stripJPEGMetadata(raw []byte) ([]byte, error) {
	reader := bytes.NewReader(raw)
	if err := validateRasterDimensions(reader); err != nil {
		return nil, err
	}

	orient := readJPEGExifOrientation(raw)

	img, err := jpeg.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("invalid JPEG image data: %w", err)
	}

	if orient > 1 && orient <= 8 {
		img = applyOrientation(img, orient)
	}

	var out bytes.Buffer
	if err := jpeg.Encode(&out, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("failed to re-encode clean JPEG: %w", err)
	}
	return out.Bytes(), nil
}

func stripPNGMetadata(raw []byte) ([]byte, error) {
	reader := bytes.NewReader(raw)
	if err := validateRasterDimensions(reader); err != nil {
		return nil, err
	}

	img, err := png.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("invalid PNG image data: %w", err)
	}

	var out bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.DefaultCompression}
	if err := enc.Encode(&out, img); err != nil {
		return nil, fmt.Errorf("failed to re-encode clean PNG: %w", err)
	}
	return out.Bytes(), nil
}

func stripGIFMetadata(raw []byte) ([]byte, error) {
	reader := bytes.NewReader(raw)
	if err := validateRasterDimensions(reader); err != nil {
		return nil, err
	}

	g, err := gif.DecodeAll(reader)
	if err != nil {
		return nil, fmt.Errorf("invalid GIF image data: %w", err)
	}

	// gif.DecodeAll/EncodeAll discards all GIF Comment (0xFE) and XMP Application (0xFF) extensions
	var out bytes.Buffer
	if err := gif.EncodeAll(&out, g); err != nil {
		return nil, fmt.Errorf("failed to re-encode clean GIF: %w", err)
	}
	return out.Bytes(), nil
}

// StripWebPMetadata parses a RIFF/WEBP container, validates chunk structure and image
// dimensions, removes EXIF, XMP, ICCP, unknown metadata chunks, and trailing data,
// clears VP8X metadata flags, and rebuilds a clean RIFF WebP binary.
func StripWebPMetadata(raw []byte) ([]byte, error) {
	if len(raw) < 20 || len(raw) > MaxImageBytes {
		return nil, errors.New("invalid WebP file size")
	}
	if !bytes.Equal(raw[:4], []byte("RIFF")) || !bytes.Equal(raw[8:12], []byte("WEBP")) {
		return nil, errors.New("invalid WebP RIFF signature")
	}

	declaredSize := int(binary.LittleEndian.Uint32(raw[4:8]))
	if declaredSize < 4 {
		return nil, errors.New("corrupt WebP container size")
	}
	limit := 8 + declaredSize
	if limit > len(raw) {
		// Allow slight tolerance only if entire buffer is valid chunks; otherwise reject truncated files
		limit = len(raw)
	}

	var chunksOut bytes.Buffer
	hasBitstream := false
	offset := 12

	for offset < limit {
		if offset+8 > limit {
			return nil, errors.New("malformed WebP chunk header")
		}
		fourCC := string(raw[offset : offset+4])
		chunkSize := int(binary.LittleEndian.Uint32(raw[offset+4 : offset+8]))
		if chunkSize < 0 || offset+8+chunkSize > limit {
			return nil, errors.New("truncated WebP chunk payload")
		}

		payload := raw[offset+8 : offset+8+chunkSize]
		nextOffset := offset + 8 + chunkSize
		if chunkSize%2 == 1 && nextOffset < limit {
			nextOffset++
		}

		switch fourCC {
		case "VP8X":
			if chunkSize != 10 {
				return nil, errors.New("invalid VP8X header length")
			}
			// Canvas width/height in VP8X are 24-bit little-endian (stored as width-1, height-1)
			w := 1 + (int(payload[4]) | int(payload[5])<<8 | int(payload[6])<<16)
			h := 1 + (int(payload[7]) | int(payload[8])<<8 | int(payload[9])<<16)
			if w > MaxImageDimension || h > MaxImageDimension || int64(w)*int64(h) > MaxImagePixels {
				return nil, fmt.Errorf("WebP dimensions (%dx%d) exceed safety limit", w, h)
			}
			cleanVP8X := make([]byte, 10)
			copy(cleanVP8X, payload)
			// Clear ICCP (0x20), EXIF (0x08), and XMP (0x04) flags
			cleanVP8X[0] &^= (0x20 | 0x08 | 0x04)
			writeRIFFChunk(&chunksOut, "VP8X", cleanVP8X)

		case "VP8 ":
			if chunkSize < 10 {
				return nil, errors.New("corrupt VP8 bitstream")
			}
			hasBitstream = true
			writeRIFFChunk(&chunksOut, "VP8 ", payload)

		case "VP8L":
			if chunkSize < 5 {
				return nil, errors.New("corrupt VP8L bitstream")
			}
			hasBitstream = true
			writeRIFFChunk(&chunksOut, "VP8L", payload)

		case "ALPH", "ANIM":
			writeRIFFChunk(&chunksOut, fourCC, payload)

		case "ANMF":
			cleanANMF, err := sanitizeWebPANMF(payload)
			if err != nil {
				return nil, err
			}
			hasBitstream = true
			writeRIFFChunk(&chunksOut, "ANMF", cleanANMF)

		default:
			// Drop EXIF, XMP , ICCP, and any non-standard metadata chunks
		}

		offset = nextOffset
	}

	if !hasBitstream {
		return nil, errors.New("WebP file contains no valid image bitstream")
	}

	bodyLen := 4 + chunksOut.Len()
	if bodyLen < 4 || bodyLen > MaxImageBytes {
		return nil, errors.New("rebuilt WebP exceeds maximum allowed size")
	}

	var out bytes.Buffer
	out.Grow(8 + bodyLen)
	out.WriteString("RIFF")
	var sizeBuf [4]byte
	// #nosec G115 -- bodyLen is strictly bounded between 4 and MaxImageBytes (25 MB) above
	binary.LittleEndian.PutUint32(sizeBuf[:], uint32(bodyLen))
	out.Write(sizeBuf[:])
	out.WriteString("WEBP")
	out.Write(chunksOut.Bytes())
	return out.Bytes(), nil
}

func sanitizeWebPANMF(payload []byte) ([]byte, error) {
	if len(payload) < 16 {
		return nil, errors.New("malformed WebP ANMF frame")
	}
	var out bytes.Buffer
	out.Write(payload[:16])

	offset := 16
	hasFrameData := false
	for offset < len(payload) {
		if offset+8 > len(payload) {
			return nil, errors.New("malformed WebP ANMF subchunk")
		}
		subID := string(payload[offset : offset+4])
		subSize := int(binary.LittleEndian.Uint32(payload[offset+4 : offset+8]))
		if subSize < 0 || offset+8+subSize > len(payload) {
			return nil, errors.New("truncated WebP ANMF subchunk")
		}
		subPayload := payload[offset+8 : offset+8+subSize]
		next := offset + 8 + subSize
		if subSize%2 == 1 && next < len(payload) {
			next++
		}

		switch subID {
		case "ALPH":
			writeRIFFChunk(&out, subID, subPayload)
		case "VP8 ", "VP8L":
			hasFrameData = true
			writeRIFFChunk(&out, subID, subPayload)
		}
		offset = next
	}

	if !hasFrameData {
		return nil, errors.New("WebP animation frame missing bitstream")
	}
	return out.Bytes(), nil
}

func writeRIFFChunk(buf *bytes.Buffer, fourCC string, payload []byte) {
	buf.WriteString(fourCC[:4])
	var sz [4]byte
	// #nosec G115 -- payload comes from slice bounded by MaxImageBytes (25 MB)
	binary.LittleEndian.PutUint32(sz[:], uint32(len(payload)))
	buf.Write(sz[:])
	buf.Write(payload)
	if len(payload)%2 == 1 {
		buf.WriteByte(0)
	}
}

// readJPEGExifOrientation extracts the EXIF Orientation tag (1..8) before metadata is stripped
// so portrait camera photos remain properly oriented after re-encoding.
func readJPEGExifOrientation(raw []byte) int {
	if len(raw) < 4 || raw[0] != 0xFF || raw[1] != 0xD8 {
		return 1
	}
	pos := 2
	for pos+4 <= len(raw) {
		if raw[pos] != 0xFF {
			break
		}
		for pos < len(raw) && raw[pos] == 0xFF {
			pos++
		}
		if pos >= len(raw) {
			break
		}
		marker := raw[pos]
		pos++

		// Standalone markers or start of compressed scan
		if marker == 0xD9 || marker == 0xDA {
			break
		}
		if marker == 0x01 || (marker >= 0xD0 && marker <= 0xD7) {
			continue
		}
		if pos+2 > len(raw) {
			break
		}
		segLen := int(binary.BigEndian.Uint16(raw[pos : pos+2]))
		if segLen < 2 || pos+segLen > len(raw) {
			break
		}

		if marker == 0xE1 && segLen >= 16 && bytes.Equal(raw[pos+2:pos+8], []byte("Exif\x00\x00")) {
			tiff := raw[pos+8 : pos+segLen]
			if orient := parseTIFFOrientation(tiff); orient >= 1 && orient <= 8 {
				return orient
			}
		}
		pos += segLen
	}
	return 1
}

func parseTIFFOrientation(tiff []byte) int {
	if len(tiff) < 8 {
		return 1
	}
	var bo binary.ByteOrder
	switch string(tiff[:2]) {
	case "II":
		bo = binary.LittleEndian
	case "MM":
		bo = binary.BigEndian
	default:
		return 1
	}
	if bo.Uint16(tiff[2:4]) != 0x002A {
		return 1
	}
	ifdOffset := int(bo.Uint32(tiff[4:8]))
	if ifdOffset < 8 || ifdOffset+2 > len(tiff) {
		return 1
	}
	numEntries := int(bo.Uint16(tiff[ifdOffset : ifdOffset+2]))
	for i := 0; i < numEntries; i++ {
		entry := ifdOffset + 2 + i*12
		if entry+12 > len(tiff) {
			break
		}
		tag := bo.Uint16(tiff[entry : entry+2])
		if tag == 0x0112 {
			val := int(bo.Uint16(tiff[entry+8 : entry+10]))
			if val >= 1 && val <= 8 {
				return val
			}
			return 1
		}
	}
	return 1
}

func applyOrientation(src image.Image, orient int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return src
	}

	switch orient {
	case 2: // Flip horizontal
		dst := image.NewNRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(w-1-x, y, src.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return dst
	case 3: // Rotate 180
		dst := image.NewNRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(w-1-x, h-1-y, src.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return dst
	case 4: // Flip vertical
		dst := image.NewNRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(x, h-1-y, src.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return dst
	case 5: // Transpose
		dst := image.NewNRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(y, x, src.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return dst
	case 6: // Rotate 90 CW
		dst := image.NewNRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(h-1-y, x, src.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return dst
	case 7: // Transverse
		dst := image.NewNRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(h-1-y, w-1-x, src.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return dst
	case 8: // Rotate 270 CW
		dst := image.NewNRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(y, w-1-x, src.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return dst
	default:
		return src
	}
}
