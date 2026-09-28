package worker

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool_Execution(t *testing.T) {
	p := NewPool(4, 32, 2*time.Second)
	defer p.Shutdown(1 * time.Second)

	var counter int32
	const jobs = 20

	for i := 0; i < jobs; i++ {
		// Fiber buffer safety demonstration: clone any string before passing to closure
		rawPath := "/blog/high-performance-go"
		clonedPath := strings.Clone(rawPath)

		p.Enqueue(func(ctx context.Context) {
			if len(clonedPath) > 0 {
				atomic.AddInt32(&counter, 1)
			}
		})
	}

	time.Sleep(100 * time.Millisecond)

	if count := atomic.LoadInt32(&counter); count != jobs {
		t.Errorf("Expected %d completed jobs, got %d", jobs, count)
	}
}

func TestPool_SaturationDrop(t *testing.T) {
	// Tiny queue of 2
	p := NewPool(1, 2, 1*time.Second)
	defer p.Shutdown(500 * time.Millisecond)

	blocker := make(chan struct{})

	// Fill queue with blocking tasks
	p.Enqueue(func(ctx context.Context) {
		<-blocker
	})
	p.Enqueue(func(ctx context.Context) {
		<-blocker
	})
	p.Enqueue(func(ctx context.Context) {
		<-blocker
	})

	// 4th task should be dropped non-blockingly
	enqueued := p.Enqueue(func(ctx context.Context) {})
	if enqueued {
		t.Errorf("Expected task to be dropped when queue is saturated, but Enqueue returned true")
	}

	close(blocker)
}
