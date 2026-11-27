package services

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/JRoetscyber/my_website/go_app/internal/utils"
)

type WebPConversionResult struct {
	OriginalFilename string  `json:"original_filename"`
	WebPFilename     string  `json:"webp_filename"`
	DownloadURL      string  `json:"download_url"`
	InputSizeKB      float64 `json:"input_size_kb"`
	OutputSizeKB     float64 `json:"output_size_kb"`
	SavingsPercent   float64 `json:"savings_percent"`
	Quality          string  `json:"quality"`
	Method           string  `json:"method"`
}

// ConvertToWebP converts an image file (JPG, PNG, BMP) to WebP format
func ConvertToWebP(inputPath, outputPath, quality, method string) (*WebPConversionResult, error) {
	cleanIn := filepath.Clean(inputPath)
	cleanOut := filepath.Clean(outputPath)
	if cleanIn == "" || cleanOut == "" || strings.HasPrefix(cleanIn, "-") || strings.HasPrefix(cleanOut, "-") {
		return nil, errors.New("invalid input or output path")
	}

	q := strings.TrimSpace(quality)
	if q == "" || strings.EqualFold(q, "auto") {
		q = "80"
	} else if val, err := strconv.Atoi(q); err != nil || val < 0 || val > 100 {
		q = "80"
	}

	m := strings.TrimSpace(method)
	if val, err := strconv.Atoi(m); err != nil || val < 0 || val > 6 {
		m = "6"
	}

	// 1. Locate converter executable (Custom C++ tool or cwebp)
	converterBin := findConverterBinary()

	var cmd *exec.Cmd
	if strings.Contains(converterBin, "converter") {
		// Custom C++ converter: ./converter <input> <output> [quality] [method]
		// #nosec G204 -- converterBin is resolved from fixed allow-listed paths and args/paths are strictly validated above
		cmd = exec.Command(converterBin, cleanIn, cleanOut, q, m)
	} else if strings.Contains(converterBin, "cwebp") {
		// Google cwebp fallback: cwebp -metadata none -q <quality> -m <method> <input> -o <output>
		// #nosec G204 -- binary is fixed literal "cwebp" and q/m/paths are strictly validated above
		cmd = exec.Command("cwebp", "-metadata", "none", "-q", q, "-m", m, cleanIn, "-o", cleanOut)
	} else {
		return nil, errors.New("no WebP encoder found (neither custom converter nor cwebp is available on host)")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("webp conversion failed: %v (output: %s)", err, string(output))
	}

	// Guarantee 100% of EXIF/XMP/ICCP chunks are stripped from the output WebP container using os.Root containment
	if outRoot, rootErr := os.OpenRoot(filepath.Dir(cleanOut)); rootErr == nil {
		outName := filepath.Base(cleanOut)
		if rawOut, readErr := outRoot.ReadFile(outName); readErr == nil {
			if strippedOut, stripErr := utils.StripWebPMetadata(rawOut); stripErr == nil {
				_ = outRoot.WriteFile(outName, strippedOut, 0600)
			}
		}
		_ = outRoot.Close()
	}

	// Verify output exists
	outInfo, err := os.Stat(cleanOut)
	if err != nil || outInfo.Size() == 0 {
		return nil, errors.New("converted webp file is missing or 0 bytes")
	}

	inInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil, err
	}

	inputSize := inInfo.Size()
	outputSize := outInfo.Size()

	savings := 0.0
	if inputSize > 0 {
		savings = math.Round((1.0-float64(outputSize)/float64(inputSize))*10000.0) / 100.0
	}

	return &WebPConversionResult{
		OriginalFilename: filepath.Base(inputPath),
		WebPFilename:     filepath.Base(outputPath),
		DownloadURL:      "/static/uploads/webp/" + filepath.Base(outputPath),
		InputSizeKB:      math.Round(float64(inputSize)/1024.0*10.0) / 10.0,
		OutputSizeKB:     math.Round(float64(outputSize)/1024.0*10.0) / 10.0,
		SavingsPercent:   savings,
		Quality:          quality,
		Method:           method,
	}, nil
}

func findConverterBinary() string {
	candidates := []string{
		"/app/converter",
		"./converter",
		"../converter",
		"converter",
		"./webp_converter/converter",
		"../webp_converter/converter",
	}

	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return c
		}
	}

	if path, err := exec.LookPath("cwebp"); err == nil {
		return path
	}

	return ""
}
