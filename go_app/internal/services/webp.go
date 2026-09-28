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
	if quality == "" {
		quality = "80"
	}
	if method == "" {
		method = "6"
	}

	// 1. Locate converter executable (Custom C++ tool or cwebp)
	converterBin := findConverterBinary()

	var cmd *exec.Cmd
	if strings.Contains(converterBin, "converter") {
		// Custom C++ converter: ./converter <input> <output> [quality] [method]
		cmd = exec.Command(converterBin, inputPath, outputPath, quality, method)
	} else if strings.Contains(converterBin, "cwebp") {
		// Google cwebp fallback: cwebp -q <quality> -m <method> <input> -o <output>
		q := quality
		if strings.EqualFold(q, "auto") {
			q = "80"
		}
		// ensure quality is integer 0-100
		if val, err := strconv.Atoi(q); err != nil || val < 0 || val > 100 {
			q = "80"
		}
		m := method
		if val, err := strconv.Atoi(m); err != nil || val < 0 || val > 6 {
			m = "6"
		}
		cmd = exec.Command("cwebp", "-q", q, "-m", m, inputPath, "-o", outputPath)
	} else {
		return nil, errors.New("no WebP encoder found (neither custom converter nor cwebp is available on host)")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("webp conversion failed: %v (output: %s)", err, string(output))
	}

	// Verify output exists
	outInfo, err := os.Stat(outputPath)
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
