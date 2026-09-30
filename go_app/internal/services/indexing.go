package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/config"
)

// IndexNowPayload represents the JSON body expected by the IndexNow API
type IndexNowPayload struct {
	Host        string   `json:"host"`
	Key         string   `json:"key"`
	KeyLocation string   `json:"keyLocation,omitempty"`
	URLList     []string `json:"urlList"`
}

// NotifySearchEngines broadcasts newly published/updated URLs to IndexNow (Bing, Yandex, etc.)
// and submits sitemap refresh pings. Runs safely with dedicated timeouts.
func NotifySearchEngines(cfg *config.Config, urls []string) error {
	if len(urls) == 0 {
		return nil
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://jo4.co.za"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	parsed, err := url.Parse(baseURL)
	host := "jo4.co.za"
	if err == nil && parsed.Host != "" {
		host = parsed.Host
	}

	key := cfg.IndexNowKey
	if key == "" {
		key = "8fa732e604bf44f89d3d3a089cf1dfc8"
	}

	// Format URLs to ensure absolute targets
	var validURLs []string
	for _, u := range urls {
		trimmed := strings.TrimSpace(u)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
			trimmed = baseURL + "/" + strings.TrimPrefix(trimmed, "/")
		}
		validURLs = append(validURLs, trimmed)
	}

	if len(validURLs) == 0 {
		return nil
	}

	payload := IndexNowPayload{
		Host:        host,
		Key:         key,
		KeyLocation: fmt.Sprintf("%s/%s.txt", baseURL, key),
		URLList:     validURLs,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal IndexNow payload: %w", err)
	}

	client := &http.Client{
		Timeout: 7 * time.Second,
	}

	// 1. Submit to IndexNow API (shared across Bing, Yandex, Naver, Seznam)
	indexNowEndpoints := []string{
		"https://api.indexnow.org/indexnow",
		"https://www.bing.com/indexnow",
	}

	for _, endpoint := range indexNowEndpoints {
		req, reqErr := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
		if reqErr != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("User-Agent", "JO4-Dev-AutoIndexer/1.0")

		resp, postErr := client.Do(req)
		if postErr != nil {
			log.Printf("[Indexing] Warning: Failed to submit to %s: %v", endpoint, postErr)
			continue
		}
		resp.Body.Close()
		log.Printf("[Indexing] IndexNow submission to %s returned HTTP %d for %d URLs", endpoint, resp.StatusCode, len(validURLs))
	}

	// 2. Google Sitemap Ping (fallback notification)
	sitemapURL := fmt.Sprintf("%s/sitemap.xml", baseURL)
	googlePing := fmt.Sprintf("https://www.google.com/ping?sitemap=%s", url.QueryEscape(sitemapURL))
	if pingReq, pingErr := http.NewRequest("GET", googlePing, nil); pingErr == nil {
		pingReq.Header.Set("User-Agent", "JO4-Dev-AutoIndexer/1.0")
		if pingResp, err := client.Do(pingReq); err == nil {
			pingResp.Body.Close()
			log.Printf("[Indexing] Google sitemap ping returned HTTP %d", pingResp.StatusCode)
		}
	}

	return nil
}
