package main

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
	semaphore     chan struct{}
}

func NewPool(maxWorkers, bufferSize int) *WorkerPool {
	return &WorkerPool{
		maxWorkers:    maxWorkers,
		wg:            &sync.WaitGroup{},
		workers:       NewDynamicBuffer(2, bufferSize),
		maxBufferSize: int64(bufferSize),
		semaphore:     make(chan struct{}, bufferSize),
	}
}

func (p *WorkerPool) Run() {
	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *WorkerPool) Submit(worker *Worker) {
	if p.IsFull() {
		<-p.semaphore
	}
	p.workers.Add(worker)
	defer worker.Kill()
	atomic.AddInt64(&p.Len, 1)
}

func (p *WorkerPool) worker() {
	defer p.wg.Done()
	for {
		w, ok := <-p.workers.buffer
		if !ok {
			return
		}
		w.Task(w.Ctx)
		atomic.AddInt64(&p.Len, -1)
		if p.hasSpace() {
			p.semaphore <- struct{}{}
		}
	}
}

func (p *WorkerPool) len() int64 {
	return atomic.LoadInt64(&p.Len)
}
func (p *WorkerPool) IsFull() bool {
	return atomic.LoadInt64(&p.Len) == p.maxBufferSize
}

func (p *WorkerPool) hasSpace() bool {
	return p.workers.Cap() == p.maxBufferSize && p.len() == p.maxBufferSize-1
}
func (p *WorkerPool) Wait() {
	close(p.workers.buffer)
	p.wg.Wait()
}

func main() {
	pool := NewPool(2, 20)

	pool.Run()
	for i := 1; i <= 5; i++ {
		i := i
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
			//println(pool.len(), pool.workers.Cap())
			time.Sleep(2 * time.Second)
			if ctx.Err() != nil {
				println("slow1 canceled!!!!!!!")
				return
			}
		}))
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
			//println(pool.len(), pool.workers.Cap())
			println("fast1", i)
		}))
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
			//println(pool.len(), pool.workers.Cap())
			println("fast2", i)
		}))
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
			//println(pool.len(), pool.workers.Cap())
			time.Sleep(2 * time.Second)
			println("slow2", i)
		}))
	}
	pool.Wait()
	println("end")
}
