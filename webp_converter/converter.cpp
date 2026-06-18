// ================================================================
//  converter.cpp  —  Silent WebP backend utility for Flask
//  Build:  g++ converter.cpp -o converter -lwebp -lm -fopenmp -O3
// ================================================================

#include <string>
#include <vector>
#include <cmath>
#include <cstring>
#include <sys/stat.h>
#include <algorithm>
#include <omp.h> 

#define STB_IMAGE_IMPLEMENTATION
#include "stb_image.h"

#include <webp/encode.h>
#include <webp/decode.h>

long fileSize(const std::string& path) {
    struct stat s;
    return (stat(path.c_str(), &s) == 0) ? s.st_size : -1;
}

float clamp(float v, float lo, float hi) {
    return v < lo ? lo : (v > hi ? hi : v);
}

float analyseComplexity(const unsigned char* img, int w, int h, int ch) {
    double totalVariance = 0.0;
    int    blocks        = 0;
    int    blockSize     = 16;

    #pragma omp parallel for reduction(+:totalVariance, blocks)
    for (int y = 0; y <= h - blockSize; y += blockSize) {
        for (int x = 0; x <= w - blockSize; x += blockSize) {
            double sum = 0, sumSq = 0;
            int    n   = blockSize * blockSize;

            for (int by = 0; by < blockSize; by++) {
                for (int bx = 0; bx < blockSize; bx++) {
                    int idx       = ((y + by) * w + (x + bx)) * ch;
                    float lum     = 0.299f * img[idx] + 0.587f * img[idx+1] + 0.114f * img[idx+2];
                    sum          += lum;
                    sumSq        += lum * lum;
                }
            }
            double mean      = sum / n;
            double variance  = (sumSq / n) - (mean * mean);
            totalVariance   += variance;
            blocks++;
        }
    }
    return blocks > 0 ? (float)(totalVariance / blocks) : 0.0f;
}

float smartQuality(float complexity) {
    float norm    = clamp(complexity / 4000.0f, 0.0f, 1.0f);
    return 15.0f + norm * 20.0f;   // Start extremely aggressive
}

void unsharpMask(unsigned char* img, int w, int h, int ch, float amount) {
    std::vector<float> blurred(w * h * ch);

    #pragma omp parallel for
    for (int y = 0; y < h; y++) {
        for (int x = 0; x < w; x++) {
            for (int c = 0; c < ch; c++) {
                float sum = 0;
                int   cnt = 0;
                for (int dy = -1; dy <= 1; dy++) {
                    for (int dx = -1; dx <= 1; dx++) {
                        int ny = y + dy, nx = x + dx;
                        if (ny >= 0 && ny < h && nx >= 0 && nx < w) {
                            sum += img[(ny * w + nx) * ch + c];
                            cnt++;
                        }
                    }
                }
                blurred[(y * w + x) * ch + c] = sum / cnt;
            }
        }
    }

    #pragma omp parallel for
    for (int i = 0; i < w * h * ch; i++) {
        float sharpened = img[i] + amount * (img[i] - blurred[i]);
        img[i] = (unsigned char)clamp(sharpened, 0.0f, 255.0f);
    }
}

void denoiseLight(unsigned char* img, int w, int h, int ch) {
    std::vector<unsigned char> out(w * h * ch);
    float sigmaColor = 25.0f;

    #pragma omp parallel for
    for (int y = 1; y < h - 1; y++) {
        for (int x = 1; x < w - 1; x++) {
            for (int c = 0; c < ch; c++) {
                float centerVal = img[(y * w + x) * ch + c];
                float sum       = 0;
                float weightSum = 0;

                for (int dy = -1; dy <= 1; dy++) {
                    for (int dx = -1; dx <= 1; dx++) {
                        float neighborVal = img[((y+dy) * w + (x+dx)) * ch + c];
                        float diff        = centerVal - neighborVal;
                        float weight      = expf(-(diff * diff) / (2.0f * sigmaColor * sigmaColor));
                        sum              += weight * neighborVal;
                        weightSum        += weight;
                    }
                }
                out[(y * w + x) * ch + c] = (unsigned char)(sum / weightSum);
            }
        }
    }

    for(int y = 1; y < h - 1; y++) {
        memcpy(&img[(y * w + 1) * ch], &out[(y * w + 1) * ch], (w - 2) * ch);
    }
}

double computeSSIM(const unsigned char* a, const unsigned char* b, int w, int h) {
    double ssimTotal = 0;
    int    windows   = 0;
    int    win       = 8;
    const double C1 = 6.5025,  C2 = 58.5225;

    #pragma omp parallel for reduction(+:ssimTotal, windows)
    for (int y = 0; y <= h - win; y += win) {
        for (int x = 0; x <= w - win; x += win) {
            double muA = 0, muB = 0;
            int    n   = win * win;

            for (int dy = 0; dy < win; dy++) {
                for (int dx = 0; dx < win; dx++) {
                    int idx = ((y+dy)*w + (x+dx)) * 3;
                    muA += 0.299*a[idx] + 0.587*a[idx+1] + 0.114*a[idx+2];
                    muB += 0.299*b[idx] + 0.587*b[idx+1] + 0.114*b[idx+2];
                }
            }
            muA /= n; muB /= n;

            double sigA = 0, sigB = 0, sigAB = 0;
            for (int dy = 0; dy < win; dy++) {
                for (int dx = 0; dx < win; dx++) {
                    int    idx = ((y+dy)*w + (x+dx)) * 3;
                    double lA  = 0.299*a[idx] + 0.587*a[idx+1] + 0.114*a[idx+2] - muA;
                    double lB  = 0.299*b[idx] + 0.587*b[idx+1] + 0.114*b[idx+2] - muB;
                    sigA  += lA * lA;
                    sigB  += lB * lB;
                    sigAB += lA * lB;
                }
            }
            sigA  = sqrt(sigA  / n);
            sigB  = sqrt(sigB  / n);
            sigAB =      sigAB / n;

            double num = (2*muA*muB + C1) * (2*sigAB + C2);
            double den = (muA*muA + muB*muB + C1) * (sigA*sigA + sigB*sigB + C2);
            ssimTotal += num / den;
            windows++;
        }
    }
    return windows > 0 ? ssimTotal / windows : 0.0;
}

bool encodeWebP(const unsigned char* pixels, int w, int h,
                float quality, int method, WebPMemoryWriter& writer) {
    WebPConfig config;
    if (!WebPConfigInit(&config)) return false;

    config.quality            = quality;
    config.method             = method;
    config.pass               = 6;        
    config.autofilter         = 1;
    config.filter_sharpness   = 2;        
    config.preprocessing      = 0;        
    config.segments           = 4;
    config.sns_strength       = 0;        
    config.filter_strength    = 35;       
    config.filter_type        = 1;
    config.thread_level       = 1;        

    if (!WebPValidateConfig(&config)) return false;

    WebPPicture picture;
    if (!WebPPictureInit(&picture)) return false;

    picture.width      = w;
    picture.height     = h;
    picture.use_argb   = 1;

    if (!WebPPictureImportRGB(&picture, pixels, w * 3)) {
        WebPPictureFree(&picture);
        return false;
    }

    WebPMemoryWriterInit(&writer);
    picture.writer     = WebPMemoryWrite;
    picture.custom_ptr = &writer;

    bool ok = WebPEncode(&config, &picture);
    WebPPictureFree(&picture);
    return ok;
}

int main(int argc, char* argv[]) {
    if (argc < 3) return 1;

    std::string inputFile  = argv[1];
    std::string outputFile = argv[2];
    bool        autoQual   = true;
    float       quality    = 0;
    int         method     = 6;

    if (argc >= 4) {
        std::string qarg = argv[3];
        if (qarg != "auto") { autoQual = false; quality = clamp(std::stof(qarg), 0, 100); }
    }
    if (argc >= 5) method = std::max(0, std::min(6, atoi(argv[4])));

    int width, height, channels;
    unsigned char* raw = stbi_load(inputFile.c_str(), &width, &height, &channels, 3);
    if (!raw) return 1; // Fails silently if image is broken

    float complexity = analyseComplexity(raw, width, height, 3);
    if (autoQual) quality = smartQuality(complexity);

    denoiseLight(raw, width, height, 3);

    float sharpAmount = clamp(0.40f - (complexity / 8000.0f), 0.10f, 0.40f);
    unsharpMask(raw, width, height, 3, sharpAmount);

    WebPMemoryWriter writer;
    if (!encodeWebP(raw, width, height, quality, method, writer)) return 1;

    int decW, decH;
    unsigned char* decoded = WebPDecodeRGB(writer.mem, writer.size, &decW, &decH);

    if (decoded) {
        double ssim = computeSSIM(raw, decoded, width, height);
        WebPFree(decoded);

        int retries = 0;
        while (ssim < 0.945 && quality < 85 && retries < 4) {
            WebPMemoryWriterClear(&writer);
            quality += 4; 
            encodeWebP(raw, width, height, quality, method, writer);
            
            decoded = WebPDecodeRGB(writer.mem, writer.size, &decW, &decH);
            if (decoded) {
                ssim = computeSSIM(raw, decoded, width, height);
                WebPFree(decoded);
            }
            retries++;
        }
    }

    FILE* outFile = fopen(outputFile.c_str(), "wb");
    if (outFile) {
        fwrite(writer.mem, 1, writer.size, outFile);
        fclose(outFile);
    } else {
        return 1;
    }

    WebPMemoryWriterClear(&writer);
    stbi_image_free(raw);
    return 0; // Success
}
