package workerpool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Worker struct {
	Ctx     context.Context
	Cancel  context.CancelFunc
	TimeOut *time.Duration
	Task    func(ctx context.Context)
}

func NewWorker(ctx context.Context, task func(ctx context.Context), timeOut ...time.Duration) *Worker {
	var (
		kill    context.CancelFunc
		timeout *time.Duration
	)

	if len(timeOut) > 0 {
		timeout = &timeOut[0]
		ctx, kill = context.WithTimeout(ctx, *timeout)
	}

	return &Worker{Ctx: ctx, TimeOut: timeout, Task: task, Cancel: kill}
}

// Kill stops the worker if it has a timeout set.
func (w *Worker) Kill() {
	if w.TimeOut != nil {
		w.Cancel()
	}
}

type WorkerPool struct {
	maxWorkers    int
	workers       *DynamicBuffer
	wg            *sync.WaitGroup
	Len           int64
	maxBufferSize int64
	waitSpace     chan struct{}
}

func NewPool(maxWorkers, bufferSize int) *WorkerPool {
	return &WorkerPool{
		maxWorkers:    maxWorkers,
		wg:            &sync.WaitGroup{},
		workers:       NewDynamicBuffer(2, bufferSize),
		maxBufferSize: int64(bufferSize),
		waitSpace:     make(chan struct{}, bufferSize),
	}
}

// Run runs the WorkerPool.
//
// It creates a specified number of worker goroutines and starts executing the worker function.
// The number of worker goroutines is determined by the maxWorkers field of the WorkerPool.
// This function uses the sync.WaitGroup wg to wait for all worker goroutines to finish.
// The worker function is executed concurrently in each goroutine.
func (p *WorkerPool) Run() {
	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// Submit submits a worker to the WorkerPool.
//
// It checks if the WorkerPool is full and blocks until space is available.
// It then adds the worker to the WorkerPool and increments the length.
//
// Example
//
//	workerpool := workerpool.NewPool(2, 20)
//	worker := workerpool.NewWorker(context.Background(), func(ctx context.Context) {})
//	workerpool.Submit(worker)
//
//	// OR
//
//	workerpool.Submit(workerpool.NewWorker(context.Background(), func(ctx context.Context) {}))
//
//	// WITH TIMEOUT
//
//	workerpool.Submit(workerpool.NewWorker(context.Background(),
//	  func(ctx context.Context) {}, 1*time.Second))
func (p *WorkerPool) Submit(worker *Worker) {
	if p.IsFull() {
		<-p.waitSpace
	}

	if ok := p.workers.Add(worker); ok {
		atomic.AddInt64(&p.Len, 1)
	}
}

// worker is a function that represents a worker in the WorkerPool.
//
// It reads tasks from the buffer and executes them using the worker's task function.
// The worker's context is passed as a parameter to the task function.
// After executing the task, the worker is killed.
// The length of the WorkerPool is decremented by 1.
// If there is space available in the WorkerPool, a signal is sent to the waitSpace channel.
func (p *WorkerPool) worker() {
	defer p.wg.Done()
	for {
		w, ok := <-p.workers.buffer
		if !ok {
			return
		}
		func(w *Worker) {
			defer w.Kill()
			w.Task(w.Ctx)
		}(w)

		atomic.AddInt64(&p.Len, -1)
		if p.isSpaceAvailable() {
			p.waitSpace <- struct{}{}
		}
	}
}

// len returns the current length of the WorkerPool.
func (p *WorkerPool) len() int64 {
	return atomic.LoadInt64(&p.Len)
}

// IsFull checks if the worker pool is full.
//
// It returns a boolean value indicating whether the worker pool is full or not.
func (p *WorkerPool) IsFull() bool {
	return atomic.LoadInt64(&p.Len) == p.maxBufferSize
}

// isSpaceAvailable checks if there is space available in the worker pool.
//
// It returns a boolean value indicating whether there is space available or not.
func (p *WorkerPool) isSpaceAvailable() bool {
	return p.workers.Cap() == p.maxBufferSize && p.len() == p.maxBufferSize-1
}
// Wait blocks until all workers in the worker pool have completed their tasks.
func (p *WorkerPool) Wait() {
	close(p.workers.buffer)
	p.wg.Wait()
}
