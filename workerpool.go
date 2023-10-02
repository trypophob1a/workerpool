package workerpool

import (
	"context"
	"sync"

	"github.com/trypophob1a/workerpool/queue"
)

type Result[T any] struct {
	Ok    T
	Error error
}

type Task[T any] func() Result[T]

type WorkerPool[T any] struct {
	ctx     context.Context
	wc      int
	tasks   *queue.Queue[Task[T]]
	wg      *sync.WaitGroup
	results chan Result[T]
}

func New[T any](ctx context.Context, workerCount int) *WorkerPool[T] {
	return &WorkerPool[T]{
		ctx:     ctx,
		wc:      workerCount,
		tasks:   queue.NewQueue[Task[T]](),
		wg:      &sync.WaitGroup{},
		results: make(chan Result[T], workerCount),
	}
}

func (p *WorkerPool[T]) Run() {
	for i := 0; i < p.wc; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.drainQueue()
		}()
	}
}

func (p *WorkerPool[T]) Add(task Task[T]) {
	p.tasks.Enqueue(task)
}

func (p *WorkerPool[T]) Wait(callback func(result Result[T])) {
	go func() {
		p.wg.Wait()
		close(p.results)
	}()
	for r := range p.results {
		callback(r)
	}
}

func (p *WorkerPool[T]) drainQueue() {
	for {
		if p.ctx.Err() != nil {
			return
		}

		task, ok := p.tasks.Dequeue()
		if !ok {
			return
		}
		p.results <- task()
	}
}
