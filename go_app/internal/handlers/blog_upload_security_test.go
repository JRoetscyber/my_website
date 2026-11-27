package handlers

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/config"
	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/JRoetscyber/my_website/go_app/internal/services"
	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Helper to create a genuine 1x1 PNG byte slice
func createValidPNGBytes() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 13, G: 148, B: 136, A: 255})
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func setupTestAdminHandler(t *testing.T) (*fiber.App, *AdminHandler, string) {
	tempDir, err := os.MkdirTemp("", "test_uploads_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	_ = os.Setenv("STATIC_DIR", tempDir)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	_ = db.AutoMigrate(&models.BlogPost{}, &models.Project{})

	cfg := &config.Config{
		BaseURL:     "https://jo4.co.za",
		IndexNowKey: "test_key",
	}

	handler := NewAdminHandler(db, cfg, nil, nil)

	app := fiber.New()
	app.Post("/admin/api/upload-image", handler.UploadBlogImage)
	app.Post("/admin/blogs", handler.SaveBlogPost)
	app.Post("/admin/add_project", handler.SaveProject)

	return app, handler, tempDir
}

func TestDetectContentTypeOnWebP(t *testing.T) {
	data, err := os.ReadFile("../../../static/assets/logo.webp")
	if err != nil {
		data, err = os.ReadFile("../../static/assets/logo.webp")
	}
	if err != nil {
		t.Fatalf("failed to read logo.webp: %v", err)
	}
	sniffed := http.DetectContentType(data[:512])
	t.Logf("Sniffed MIME for real WebP logo: %s", sniffed)
}

func TestUploadBlogImage_ValidImage(t *testing.T) {
	app, _, tempDir := setupTestAdminHandler(t)
	defer os.RemoveAll(tempDir)

	pngBytes := createValidPNGBytes()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "my_diagram.png")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	_, _ = part.Write(pngBytes)
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/admin/api/upload-image", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var res map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&res)

	if res["status"] != "success" {
		t.Errorf("expected success status, got %v", res["status"])
	}

	url, _ := res["url"].(string)
	if !strings.HasPrefix(url, "/static/uploads/blog_img_") || !strings.HasSuffix(url, ".png") {
		t.Errorf("expected generated secure url with prefix blog_img_ and .png, got %s", url)
	}

	markdown, _ := res["markdown"].(string)
	if !strings.HasPrefix(markdown, "![") || !strings.Contains(markdown, url) {
		t.Errorf("expected markdown tag containing %s, got %s", url, markdown)
	}
}

func TestUploadBlogImage_RejectsSVGScriptInjection(t *testing.T) {
	app, _, tempDir := setupTestAdminHandler(t)
	defer os.RemoveAll(tempDir)

	maliciousSVG := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert('xss')</script></svg>`)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "vector_attack.svg")
	_, _ = part.Write(maliciousSVG)
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/admin/api/upload-image", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for SVG upload, got %d", resp.StatusCode)
	}
}

func TestUploadBlogImage_RejectsDisguisedExecutableOrHTML(t *testing.T) {
	app, _, tempDir := setupTestAdminHandler(t)
	defer os.RemoveAll(tempDir)

	// File named .png but containing HTML
	fakePNG := []byte(`<!DOCTYPE html><html><body><h1>Injected</h1></body></html>`)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "fake_image.png")
	_, _ = part.Write(fakePNG)
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/admin/api/upload-image", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for MIME-mismatched content, got %d", resp.StatusCode)
	}
}

func TestUploadBlogImage_PathTraversalAndIDORProtection(t *testing.T) {
	app, _, tempDir := setupTestAdminHandler(t)
	defer os.RemoveAll(tempDir)

	pngBytes := createValidPNGBytes()

	// Malicious traversal filename attempt
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "../../../../etc/passwd.png")
	_, _ = part.Write(pngBytes)
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/admin/api/upload-image", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK because filename is sanitized and randomized, got %d", resp.StatusCode)
	}

	var res map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&res)

	filename, _ := res["filename"].(string)
	if strings.Contains(filename, "..") || strings.Contains(filename, "passwd") {
		t.Errorf("IDOR/Traversal leak: client filename was preserved in storage: %s", filename)
	}

	// Verify file was saved safely inside uploadDir only
	uploadDir := filepath.Join(tempDir, "uploads")
	expectedPath := filepath.Join(uploadDir, filename)
	if _, statErr := os.Stat(expectedPath); statErr != nil {
		t.Errorf("expected saved file at %s, error: %v", expectedPath, statErr)
	}
}

func TestIndexNow_CooldownThrottling(t *testing.T) {
	services.ResetIndexCooldownCache()
	services.SetIndexCooldown(500 * time.Millisecond)

	cfg := &config.Config{
		BaseURL:     "https://jo4.co.za",
		IndexNowKey: "test_key",
	}

	urls := []string{"https://jo4.co.za/blog/test-article"}

	// First call succeeds / records timestamp
	err1 := services.NotifySearchEngines(cfg, urls)
	if err1 != nil {
		t.Fatalf("first call failed: %v", err1)
	}

	// Second rapid call within cooldown should skip/throttle without error
	err2 := services.NotifySearchEngines(cfg, urls)
	if err2 != nil {
		t.Fatalf("second call within cooldown failed: %v", err2)
	}

	// Wait for cooldown to expire
	time.Sleep(550 * time.Millisecond)

	// Third call after cooldown expires should proceed
	err3 := services.NotifySearchEngines(cfg, urls)
	if err3 != nil {
		t.Fatalf("third call after cooldown failed: %v", err3)
	}
}

func TestUploadBlogImage_StripsEXIFAndXMPOnDisk(t *testing.T) {
	app, _, tempDir := setupTestAdminHandler(t)
	defer os.RemoveAll(tempDir)

	// Build a JPEG with embedded APP1 XMP <?xml ...> and GPS EXIF strings (previously rejected!)
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	img.Set(2, 2, color.RGBA{R: 20, G: 184, B: 166, A: 255})
	var rawJPG bytes.Buffer
	_ = jpeg.Encode(&rawJPG, img, &jpeg.Options{Quality: 92})
	b := rawJPG.Bytes()

	xmpPayload := []byte("http://ns.adobe.com/xap/1.0/\x00<?xml version=\"1.0\"?><x:xmpmeta>GPS:26.2041S,CameraSerial:XYZ987</x:xmpmeta>")
	var app1 bytes.Buffer
	app1.Write([]byte{0xFF, 0xE1})
	_ = binary.Write(&app1, binary.BigEndian, uint16(len(xmpPayload)+2))
	app1.Write(xmpPayload)

	var fullJPG bytes.Buffer
	fullJPG.Write(b[:2])
	fullJPG.Write(app1.Bytes())
	fullJPG.Write(b[2:])

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "exported_photo.jpeg")
	_, _ = part.Write(fullJPG.Bytes())
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/admin/api/upload-image", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for JPEG containing XMP/EXIF, got %d", resp.StatusCode)
	}

	var res map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&res)
	filename, _ := res["filename"].(string)
	savedBytes, err := os.ReadFile(filepath.Join(tempDir, "uploads", filename))
	if err != nil {
		t.Fatalf("failed to read saved image: %v", err)
	}

	for _, marker := range []string{"<?xml", "GPS:26.2041S", "CameraSerial:XYZ987"} {
		if bytes.Contains(savedBytes, []byte(marker)) {
			t.Errorf("saved image on disk still contains metadata %q", marker)
		}
	}
}

func TestDuplicateBlogSlugDeduplication(t *testing.T) {
	app, handler, tempDir := setupTestAdminHandler(t)
	defer os.RemoveAll(tempDir)

	for i := 0; i < 2; i++ {
		body := &bytes.Buffer{}
		w := multipart.NewWriter(body)
		_ = w.WriteField("title", "Identical Headline")
		_ = w.WriteField("content", "Body content")
		_ = w.WriteField("status", "draft")
		_ = w.Close()

		req := httptest.NewRequest("POST", "/admin/blogs", body)
		req.Header.Set("Content-Type", w.FormDataContentType())
		req.Header.Set("Accept", "application/json")

		resp, err := app.Test(req, 5000)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("post #%d failed: err=%v status=%d", i+1, err, resp.StatusCode)
		}
	}

	var count int64
	handler.DB.Model(&models.BlogPost{}).Count(&count)
	if count != 2 {
		t.Fatalf("expected 2 blog posts saved without unique slug collision, got %d", count)
	}
}
