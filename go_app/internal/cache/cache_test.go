package cache

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/config"
)

func TestCacheManager_Singleflight(t *testing.T) {
	cfg := &config.Config{
		RedisURL: "", // Test in-memory fallback
	}
	cm := NewManager(cfg)
	ctx := context.Background()

	var dbQueryCount int32
	fetcher := func(ctx context.Context) (string, error) {
		atomic.AddInt32(&dbQueryCount, 1)
		time.Sleep(50 * time.Millisecond) // simulate heavy DB query
		return "cold-data-result", nil
	}

	// Launch 20 concurrent goroutines querying the exact same cold key simultaneously
	const workers = 20
	done := make(chan string, workers)
	for i := 0; i < workers; i++ {
		go func() {
			val, _ := GetOrSet(cm, ctx, "test:entity:1:v1", 1*time.Minute, fetcher)
			done <- val
		}()
	}

	for i := 0; i < workers; i++ {
		res := <-done
		if res != "cold-data-result" {
			t.Errorf("Unexpected result: %s", res)
		}
	}

	// Singleflight MUST ensure the cold fetcher ran only ONCE despite 20 concurrent calls
	if count := atomic.LoadInt32(&dbQueryCount); count != 1 {
		t.Errorf("Expected exactly 1 DB execution due to Singleflight guard, got %d", count)
	}
}

func TestCacheManager_Invalidation(t *testing.T) {
	cfg := &config.Config{}
	cm := NewManager(cfg)
	ctx := context.Background()

	key := BlogKey("high-speed-go")
	val, err := GetOrSet(cm, ctx, key, 10*time.Minute, func(ctx context.Context) (string, error) {
		return "article-content", nil
	})
	if err != nil || val != "article-content" {
		t.Fatalf("Failed to cache: %v", err)
	}

	// Invalidate matching pattern
	cm.Invalidate(ctx, "blog:*")

	// Verify key was evicted
	_, found := cm.getRaw(ctx, key)
	if found {
		t.Errorf("Expected key to be evicted after Invalidate, but was still found")
	}
}
