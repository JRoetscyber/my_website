# WebP Converter (Pro - Pi Edition)

**Official Documentation & User Guide**

## Overview

The WebP Converter (Pro - Pi Edition) is a custom, high-performance C++ image optimization utility. It is engineered specifically for ARM-based Linux environments (like the Raspberry Pi) and optimized to compress high-contrast, graphic-heavy images (such as comic-style advertisements and vector art) for seamless integration into custom Flask backends.

The tool operates silently, communicating exclusively via system exit codes, making it the perfect background process for automated web asset pipelines.

## 1. Key Features

- **Multi-Core Processing:** Utilizes OpenMP to parallelize heavy spatial math across multiple CPU cores, dramatically reducing processing time on Raspberry Pi hardware.
- **Smart Complexity Analysis:** Reads the localized variance of an image to determine if it is a smooth photograph or a sharp graphic, automatically adjusting the baseline compression strategy.
- **SSIM Verification Loop:** Encodes the image and mathematically compares the output to the original using Structural Similarity (SSIM). If the quality drops below the visual threshold of ~0.945, it automatically bumps the quality and retries.
- **Pre-Processing Pipeline:** Includes a lightweight bilateral denoiser to remove compression artifacts before encoding, and a dynamic unsharp mask to keep fine lines crisp at lower bitrates.
- **Headless/Silent Execution:** Produces zero standard output (`stdout`) or error logs (`stderr`), communicating success (`0`) or failure (`1`) directly to the calling backend.

## 2. Installation & Compilation

### Prerequisites

The utility relies on the official WebP development libraries and standard C++ build tools. On any Debian/Ubuntu-based system (including Raspberry Pi OS), install the dependencies:

```bash
sudo apt-get update
sudo apt-get install build-essential libwebp-dev
```

### Compilation

Compile the source code using `g++`. It is critical to include the `-fopenmp` flag to enable multi-threading and the `-O3` flag for maximum execution speed.

```bash
g++ converter.cpp -o converter -lwebp -lm -fopenmp -O3
```

## 3. Command-Line Usage

The converter is executed via the terminal or triggered via a subprocess call in Python/Node.js.

### Syntax

```bash
./converter <input_file> <output_file> [quality] [method]
```

| Argument | Type | Description |
|-----------|------|-------------|
| `input_file` | Required | The path to the source image (supports JPG, PNG, BMP). |
| `output_file` | Required | The destination path for the output `.webp` file. |
| `quality` | Optional | `0` to `100` (manual override), or `auto` (default). `auto` engages the SSIM verification loop. |
| `method` | Optional | `0` (fastest) to `6` (slowest/best). Defaults to `6` for maximum file size reduction. |

### Example

```bash
./converter uploads/raw_poster.jpg static/optimized_poster.webp auto 6
```

## 4. How the Optimization Pipeline Works

When the `auto` quality flag is engaged, the binary executes a 5-step pipeline:

### 1. Load & Analyze

The `stb_image` library decodes the raw image. The script scans the pixels in 16×16 blocks to calculate local variance (complexity).

### 2. Determine Baseline

Based on complexity, it assigns a highly aggressive starting quality (usually between 15 and 35).

### 3. Pre-Process

#### Denoise

Smooths out microscopic sensor noise or minor JPEG artifacts so they aren't encoded into the new WebP, saving data.

#### Sharpen

Applies a calculated unsharp mask to ensure borders and text remain readable even at extreme compression rates.

### 4. Encode

The WebP encoder converts the image using the ARGB color space (ideal for flat graphics) with spatial noise shaping explicitly disabled to prevent wasting bits.

### 5. Verify & Adjust

The script decodes the newly created WebP in memory and compares it to the original raw pixels. If the SSIM score is below `0.945`, it:

1. Discards the WebP.
2. Increases the quality by `4`.
3. Attempts the encoding again.

This process can repeat up to 4 retries. Once the target quality threshold is met, the final file is written to disk.

## 5. Flask Backend Integration

Because the program is silent, it must be handled safely by the backend. It returns:

- `0` — The WebP file was successfully written.
- `1` — A fatal error occurred (e.g., corrupt input image, unwritable directory).

### Python Integration Example

```python
import subprocess

def optimize_image(input_path, output_path):
    try:
        # check=True ensures Python throws an error if the C++ script fails
        subprocess.run(
            ['./converter', input_path, output_path],
            check=True,
            capture_output=True
        )
        return True

    except subprocess.CalledProcessError:
        # The C++ script returned an exit code of 1
        return False
```

## Exit Codes Reference

| Exit Code | Meaning |
|------------|---------|
| `0` | Success |
| `1` | Fatal error during processing |

## Notes

- Designed for ARM-based Linux systems, including Raspberry Pi.
- Optimized for graphics, advertisements, vector art, and comic-style imagery.
- Uses automatic SSIM-based quality verification to maintain visual fidelity while maximizing compression.
- Intended for automated backend workflows and web asset optimization pipelines.