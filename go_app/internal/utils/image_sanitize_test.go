package utils

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestSanitizeJPEG_StripsEXIFAndXMPAndAppliesOrientation(t *testing.T) {
	// Create a 2x4 asymmetric image (top-left pixel red)
	src := image.NewRGBA(image.Rect(0, 0, 2, 4))
	src.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	var rawJPG bytes.Buffer
	if err := jpeg.Encode(&rawJPG, src, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	jpgBytes := rawJPG.Bytes()

	// Build a valid APP1 EXIF segment with Orientation = 6 (90 CW) + fake camera/GPS strings
	var tiff bytes.Buffer
	tiff.WriteString("MM")
	_ = binary.Write(&tiff, binary.BigEndian, uint16(0x002A))
	_ = binary.Write(&tiff, binary.BigEndian, uint32(8)) // IFD0 offset
	_ = binary.Write(&tiff, binary.BigEndian, uint16(1)) // 1 entry
	_ = binary.Write(&tiff, binary.BigEndian, uint16(0x0112))
	_ = binary.Write(&tiff, binary.BigEndian, uint16(3)) // SHORT
	_ = binary.Write(&tiff, binary.BigEndian, uint32(1)) // count 1
	_ = binary.Write(&tiff, binary.BigEndian, uint16(6)) // Orientation = 6
	_ = binary.Write(&tiff, binary.BigEndian, uint16(0))
	_ = binary.Write(&tiff, binary.BigEndian, uint32(0)) // next IFD
	tiff.WriteString("Apple iPhone 16 Pro GPSLatitude 26.2041S 28.0473E")

	exifPayload := append([]byte("Exif\x00\x00"), tiff.Bytes()...)
	var app1Exif bytes.Buffer
	app1Exif.Write([]byte{0xFF, 0xE1})
	_ = binary.Write(&app1Exif, binary.BigEndian, uint16(len(exifPayload)+2))
	app1Exif.Write(exifPayload)

	// Build an APP1 XMP segment containing <?xml ...> and <svg> namespace (previously caused false rejections!)
	xmpBody := []byte("http://ns.adobe.com/xap/1.0/\x00<?xml version=\"1.0\"?><x:xmpmeta><svg xmlns=\"http://www.w3.org/2000/svg\"/>Author: Jonathan Secret</x:xmpmeta>")
	var app1XMP bytes.Buffer
	app1XMP.Write([]byte{0xFF, 0xE1})
	_ = binary.Write(&app1XMP, binary.BigEndian, uint16(len(xmpBody)+2))
	app1XMP.Write(xmpBody)

	// Splice APP1 EXIF + APP1 XMP right after SOI (0xFF 0xD8) + append trailing polyglot bytes
	var withMeta bytes.Buffer
	withMeta.Write(jpgBytes[:2])
	withMeta.Write(app1Exif.Bytes())
	withMeta.Write(app1XMP.Bytes())
	withMeta.Write(jpgBytes[2:])
	withMeta.WriteString("<script>alert('polyglot')</script>")

	clean, ext, err := SanitizeAndStripMetadata(withMeta.Bytes(), "camera_photo.jpeg")
	if err != nil {
		t.Fatalf("expected JPEG with XMP/EXIF to succeed, got error: %v", err)
	}
	if ext != ".jpg" {
		t.Errorf("expected .jpg extension, got %s", ext)
	}

	for _, forbidden := range []string{"Exif", "iPhone", "GPSLatitude", "<?xml", "Author: Jonathan", "<script>"} {
		if bytes.Contains(clean, []byte(forbidden)) {
			t.Errorf("metadata leak: clean JPEG still contains %q", forbidden)
		}
	}

	// Verify Orientation=6 rotated 2x4 -> 4x2
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(clean))
	if err != nil {
		t.Fatalf("decode clean jpeg config: %v", err)
	}
	if cfg.Width != 4 || cfg.Height != 2 {
		t.Errorf("expected rotated dimensions 4x2, got %dx%d", cfg.Width, cfg.Height)
	}
}

func TestSanitizePNG_StripsTextEXIFAndXMPChunks(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	img.Set(1, 1, color.NRGBA{R: 13, G: 148, B: 136, A: 128})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	raw := buf.Bytes()

	// Inject a tEXt chunk and eXIf chunk right after IHDR (8-byte sig + 25-byte IHDR = offset 33)
	makePNGChunk := func(chunkType string, data []byte) []byte {
		var c bytes.Buffer
		_ = binary.Write(&c, binary.BigEndian, uint32(len(data)))
		c.WriteString(chunkType)
		c.Write(data)
		crc := crc32.NewIEEE()
		_, _ = crc.Write([]byte(chunkType))
		_, _ = crc.Write(data)
		_ = binary.Write(&c, binary.BigEndian, crc.Sum32())
		return c.Bytes()
	}

	textChunk := makePNGChunk("tEXt", []byte("Author\x00SecretLocationGPS"))
	itxtChunk := makePNGChunk("iTXt", []byte("XML:com.adobe.xmp\x00\x00\x00\x00\x00<?xml version=\"1.0\"?>SerialNumber999"))

	var injected bytes.Buffer
	injected.Write(raw[:33])
	injected.Write(textChunk)
	injected.Write(itxtChunk)
	injected.Write(raw[33:])

	clean, ext, err := SanitizeAndStripMetadata(injected.Bytes(), "screenshot.png")
	if err != nil {
		t.Fatalf("expected PNG with metadata chunks to succeed, got: %v", err)
	}
	if ext != ".png" {
		t.Errorf("expected .png, got %s", ext)
	}

	for _, forbidden := range []string{"SecretLocationGPS", "SerialNumber999", "<?xml", "iTXt", "tEXt"} {
		if bytes.Contains(clean, []byte(forbidden)) {
			t.Errorf("PNG metadata leak: clean PNG still contains %q", forbidden)
		}
	}
}

func TestStripWebPMetadata_RemovesEXIFXMPAndICCP(t *testing.T) {
	// Construct a valid extended WebP container with VP8X + VP8L + EXIF + XMP + ICCP
	var chunks bytes.Buffer
	// VP8X with ICCP(0x20) | EXIF(0x08) | XMP(0x04) flags set, 10x10 canvas
	vp8x := []byte{0x20 | 0x08 | 0x04, 0, 0, 0, 9, 0, 0, 9, 0, 0}
	writeRIFFChunk(&chunks, "VP8X", vp8x)
	writeRIFFChunk(&chunks, "ICCP", []byte("FakeICCProfileData"))
	writeRIFFChunk(&chunks, "VP8L", []byte{0x2F, 0x09, 0x40, 0x00, 0x00, 0x10})
	writeRIFFChunk(&chunks, "EXIF", []byte("Exif\x00\x00CameraSerial12345"))
	writeRIFFChunk(&chunks, "XMP ", []byte("<?xml version=\"1.0\"?>PrivateGPS"))

	var webpRaw bytes.Buffer
	webpRaw.WriteString("RIFF")
	var sz [4]byte
	binary.LittleEndian.PutUint32(sz[:], uint32(4+chunks.Len()))
	webpRaw.Write(sz[:])
	webpRaw.WriteString("WEBP")
	webpRaw.Write(chunks.Bytes())

	clean, ext, err := SanitizeAndStripMetadata(webpRaw.Bytes(), "hero.webp")
	if err != nil {
		t.Fatalf("StripWebPMetadata failed: %v", err)
	}
	if ext != ".webp" {
		t.Errorf("expected .webp, got %s", ext)
	}

	for _, forbidden := range []string{"EXIF", "XMP ", "ICCP", "CameraSerial12345", "PrivateGPS", "FakeICCProfileData"} {
		if bytes.Contains(clean, []byte(forbidden)) {
			t.Errorf("WebP metadata leak: clean WebP still contains %q", forbidden)
		}
	}

	// Verify VP8X flags byte at offset 20 has metadata bits cleared
	if clean[20]&(0x20|0x08|0x04) != 0 {
		t.Errorf("expected VP8X metadata flags cleared, got 0x%02x", clean[20])
	}
}

func TestSanitizeGIF_StripsComments(t *testing.T) {
	pal := color.Palette{color.Black, color.White}
	frame := image.NewPaletted(image.Rect(0, 0, 2, 2), pal)
	g := &gif.GIF{
		Image: []*image.Paletted{frame},
		Delay: []int{10},
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		t.Fatalf("encode gif: %v", err)
	}
	raw := buf.Bytes()

	// Inject a real GIF89a Comment Extension (0x21 0xFE <len> <comment> 0x00) before trailer 0x3B
	commentText := []byte("HiddenAuthorAndGPSComment")
	var injected bytes.Buffer
	injected.Write(raw[:len(raw)-1])
	injected.Write([]byte{0x21, 0xFE, byte(len(commentText))})
	injected.Write(commentText)
	injected.WriteByte(0x00)
	injected.WriteByte(0x3B) // GIF trailer

	clean, ext, err := SanitizeAndStripMetadata(injected.Bytes(), "anim.gif")
	if err != nil {
		t.Fatalf("sanitize gif failed: %v", err)
	}
	if ext != ".gif" {
		t.Errorf("expected .gif, got %s", ext)
	}
	if bytes.Contains(clean, commentText) {
		t.Errorf("GIF comment metadata was not stripped")
	}
}
