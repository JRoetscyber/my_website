package worker

import (
	"context"
	"log"
	"sync"
	"time"
)

// Task represents a discrete background unit of work that does NOT hold references
// to fasthttp buffers or *fiber.Ctx.
type Task func(ctx context.Context)

// Pool implements a bounded worker pool for asynchronous background operations
// (audit logging, metrics, notification dispatching, indexing) off the HTTP request path.
type Pool struct {
	maxWorkers   int
	taskQueue    chan Task
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	jobTimeout   time.Duration
	droppedTasks uint64
}

// NewPool initializes a bounded pool with a fixed number of workers and a buffered queue.
func NewPool(workers, queueSize int, jobTimeout time.Duration) *Pool {
	if workers <= 0 {
		workers = 16
	}
	if queueSize <= 0 {
		queueSize = 2048
	}
	if jobTimeout <= 0 {
		jobTimeout = 10 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	p := &Pool{
		maxWorkers: workers,
		taskQueue:  make(chan Task, queueSize),
		ctx:        ctx,
		cancel:     cancel,
		jobTimeout: jobTimeout,
	}

	p.start()
	return p
}

func (p *Pool) start() {
	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go func(workerID int) {
			defer p.wg.Done()
			for {
				select {
				case <-p.ctx.Done():
					return
				case task, ok := <-p.taskQueue:
					if !ok {
						return
					}
					p.executeTask(task)
				}
			}
		}(i)
	}
	log.Printf("[WORKER] Initialized bounded worker pool with %d workers (QueueBuffer=%d, JobTimeout=%v)", p.maxWorkers, cap(p.taskQueue), p.jobTimeout)
}

func (p *Pool) executeTask(task Task) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[WORKER] Recovered from task panic: %v", r)
		}
	}()

	taskCtx, cancel := context.WithTimeout(context.Background(), p.jobTimeout)
	defer cancel()

	task(taskCtx)
}

// Enqueue submits a task to the worker pool without blocking the HTTP request handler.
// If the buffer is full, it drops the task to protect system memory and logs a warning.
func (p *Pool) Enqueue(task Task) bool {
	select {
	case p.taskQueue <- task:
		return true
	default:
		log.Printf("[WORKER] Warning: Task queue saturated (%d/%d). Dropping background task to maintain latency.", len(p.taskQueue), cap(p.taskQueue))
		return false
	}
}

// Shutdown gracefully drains the queue or aborts on timeout.
func (p *Pool) Shutdown(timeout time.Duration) error {
	p.cancel()
	close(p.taskQueue)

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("[WORKER] Worker pool shut down cleanly.")
		return nil
	case <-time.After(timeout):
		log.Printf("[WORKER] Worker pool shutdown timed out after %v.", timeout)
		return context.DeadlineExceeded
	}
}
