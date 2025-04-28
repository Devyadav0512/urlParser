// pkg/workerpool/workerpool.go
package workerpool

import (
	"context"
	"ecommerce-crawler/internal/utils"
	"errors"
	"sync"
	"time"
)

type WorkerPool struct {
	tasks      chan *Task
	wg         sync.WaitGroup
	maxWorkers int
	timeout    time.Duration
}

func NewWorkerPool(maxWorkers int, timeout time.Duration) *WorkerPool {
	return &WorkerPool{
		tasks:      make(chan *Task, 1000),
		maxWorkers: maxWorkers,
		timeout:    timeout,
	}
}

func (wp *WorkerPool) AddTask(task *Task) {
	select {
	case wp.tasks <- task:
	default:
		// Drop task if queue is full to prevent deadlock
	}
}

func (wp *WorkerPool) Run(ctx context.Context, processFunc func(task *Task) error, logger *utils.Logger) {
	for i := 0; i < wp.maxWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, processFunc,logger)
	}
}

func (wp *WorkerPool) worker(ctx context.Context, processFunc func(task *Task) error, logger *utils.Logger) {
    defer wp.wg.Done()

    for {
        select {
        case <-ctx.Done():
            return
        case task, ok := <-wp.tasks:
            if !ok {
                return
            }

            _, cancel := context.WithTimeout(ctx, wp.timeout)
            err := processFunc(task)
            cancel()

            if err != nil {
                if errors.Is(err, context.DeadlineExceeded) {
                    logger.Warn("Task timed out", 
                        "url", task.URL,
                        "timeout", wp.timeout.String())
                } else if ctx.Err() == nil { // Only log if not cancelled
                    logger.Error("Task failed", 
                        "url", task.URL, 
                        "error", err)
                }
            }
        }
    }
}

func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
	close(wp.tasks)
}