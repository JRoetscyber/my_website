package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/config"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type memoryItem struct {
	data      []byte
	expiresAt time.Time
}

// Manager coordinates Redis caching, deterministic keys, singleflight stampede prevention,
// and in-memory fallback for high-throughput, low-latency architectures.
type Manager struct {
	client    *redis.Client
	sfGroup   singleflight.Group
	memCache  sync.Map
	available bool
}

// NewManager creates a tuned Redis cache manager. If Redis is unreachable,
// it gracefully degrades to in-memory singleflight caching to maintain 100% uptime.
func NewManager(cfg *config.Config) *Manager {
	m := &Manager{
		available: false,
	}

	if cfg.RedisURL != "" {
		opts := &redis.Options{
			Addr:         cfg.RedisURL,
			Password:     cfg.RedisPassword,
			DB:           0,
			PoolSize:     50,
			MinIdleConns: 10,
			PoolTimeout:  4 * time.Second,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
		}

		client := redis.NewClient(opts)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := client.Ping(ctx).Err(); err != nil {
			log.Printf("[CACHE] Redis at '%s' unavailable (%v). Falling back to in-memory cache-aside.", cfg.RedisURL, err)
		} else {
			m.client = client
			m.available = true
			log.Printf("[CACHE] Connected to Redis at '%s' (PoolSize=50, Cache-Aside active)", cfg.RedisURL)
		}
	} else {
		log.Printf("[CACHE] REDIS_URL not set. Running with high-speed in-memory Cache-Aside & Singleflight.")
	}

	return m
}

// IsRedisAvailable returns whether the live Redis backend is active
func (m *Manager) IsRedisAvailable() bool {
	return m.available
}

// Deterministic Cache Key Builders (Pattern: entity:id:version)
func FormatKey(entity, id, version string) string {
	return fmt.Sprintf("%s:%s:%s", entity, id, version)
}

func BlogKey(slug string) string {
	return FormatKey("blog", "slug:"+slug, "v1")
}

func BlogListKey() string {
	return FormatKey("blog", "list:all", "v1")
}

func ProjectKey(slug string) string {
	return FormatKey("project", "slug:"+slug, "v1")
}

func ProjectListKey() string {
	return FormatKey("project", "list:all", "v1")
}

func ServicesKey() string {
	return FormatKey("services", "list:all", "v1")
}

func FAQsKey() string {
	return FormatKey("faqs", "list:all", "v1")
}

func OrdersDisplayKey() string {
	return FormatKey("orders", "display:active", "v1")
}

func DashboardOverviewKey() string {
	return FormatKey("dashboard", "overview:stats", "v1")
}

// GetOrSet executes an explicit Cache-Aside read:
// 1. Checks Redis (or in-memory fallback).
// 2. On miss, uses singleflight.Group to ensure ONLY ONE worker executes the cold-start DB query.
// 3. Serializes using high-speed goccy/go-json and caches the result with the given TTL.
func GetOrSet[T any](m *Manager, ctx context.Context, key string, ttl time.Duration, fetch func(ctx context.Context) (T, error)) (T, error) {
	var zero T

	// 1. Fast path: check cache
	if cachedBytes, found := m.getRaw(ctx, key); found {
		var val T
		if err := json.Unmarshal(cachedBytes, &val); err == nil {
			return val, nil
		}
	}

	// 2. Thundering Herd Guard: Singleflight deduplicates concurrent cold misses
	v, err, _ := m.sfGroup.Do(key, func() (any, error) {
		// Double check cache within singleflight leader
		if cachedBytes, found := m.getRaw(ctx, key); found {
			var val T
			if err := json.Unmarshal(cachedBytes, &val); err == nil {
				return val, nil
			}
		}

		// Cold fetch execution
		data, err := fetch(ctx)
		if err != nil {
			return zero, err
		}

		// Serialize and store
		bytes, encErr := json.Marshal(data)
		if encErr == nil {
			m.setRaw(ctx, key, bytes, ttl)
		}

		return data, nil
	})

	if err != nil {
		return zero, err
	}

	return v.(T), nil
}

// Invalidate removes all keys matching prefix patterns (e.g. "blog:*", "project:*")
func (m *Manager) Invalidate(ctx context.Context, patterns ...string) {
	for _, pattern := range patterns {
		// Invalidate memory fallback
		m.memCache.Range(func(k, v any) bool {
			ks, ok := k.(string)
			if ok && matchPattern(pattern, ks) {
				m.memCache.Delete(ks)
			}
			return true
		})

		// Invalidate Redis if connected
		if m.available && m.client != nil {
			go func(pat string) {
				subCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				iter := m.client.Scan(subCtx, 0, pat, 0).Iterator()
				for iter.Next(subCtx) {
					_ = m.client.Del(subCtx, iter.Val()).Err()
				}
			}(pattern)
		}
	}
}

// Delete removes specific keys deterministically
func (m *Manager) Delete(ctx context.Context, keys ...string) {
	for _, k := range keys {
		m.memCache.Delete(k)
	}
	if m.available && m.client != nil {
		_ = m.client.Del(ctx, keys...).Err()
	}
}

func (m *Manager) getRaw(ctx context.Context, key string) ([]byte, bool) {
	if m.available && m.client != nil {
		b, err := m.client.Get(ctx, key).Bytes()
		if err == nil {
			return b, true
		}
	}

	// Memory fallback check
	if val, ok := m.memCache.Load(key); ok {
		item := val.(memoryItem)
		if time.Now().Before(item.expiresAt) {
			return item.data, true
		}
		m.memCache.Delete(key)
	}

	return nil, false
}

func (m *Manager) setRaw(ctx context.Context, key string, data []byte, ttl time.Duration) {
	if m.available && m.client != nil {
		_ = m.client.Set(ctx, key, data, ttl).Err()
	}

	m.memCache.Store(key, memoryItem{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	})
}

func matchPattern(pattern, key string) bool {
	if pattern == "*" {
		return true
	}
	if len(pattern) > 1 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(key) >= len(prefix) && key[:len(prefix)] == prefix
	}
	return pattern == key
}

// HTTP Cache-Control & ETag Utilities

// GenerateETag computes a strong or weak ETag from raw payload bytes
func GenerateETag(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf(`W/"%s"`, hex.EncodeToString(h[:8]))
}

// HandleETag checks the request's If-None-Match header against the computed ETag.
// If matching, it sets 304 Not Modified and returns true (indicating response is handled).
func HandleETag(c *fiber.Ctx, etag string) bool {
	clientETag := c.Get("If-None-Match")
	if clientETag != "" && clientETag == etag {
		c.Status(fiber.StatusNotModified)
		return true
	}
	return false
}

// SetCacheHeaders applies optimal browser and CDN caching policies
func SetCacheHeaders(c *fiber.Ctx, maxAge time.Duration, etag string) {
	c.Set("ETag", etag)
	c.Set("Cache-Control", fmt.Sprintf("public, max-age=%d, stale-while-revalidate=60", int(maxAge.Seconds())))
}
