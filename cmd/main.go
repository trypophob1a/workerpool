package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Result[T any] struct {
	Ok    T
	Error error
}

type Task[T any] struct {
	Result Result[T]
	Exec   func() Result[T]
}

func NewTask[T any](exec func() Result[T]) Task[T] {
	return Task[T]{Exec: exec}
}

type WorkerPool[T any] struct {
	workerCount int
	wg          *sync.WaitGroup
	ctx         context.Context
	tasks       chan Task[T]
	results     chan Result[T]
}

func NewWorkerPool[T any](ctx context.Context, workerCount int) *WorkerPool[T] {
	return &WorkerPool[T]{
		workerCount: workerCount,
		wg:          &sync.WaitGroup{},
		ctx:         ctx,
		tasks:       make(chan Task[T]),
		results:     make(chan Result[T]),
	}
}

func (p *WorkerPool[T]) Run() {
	for i := 1; i <= p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *WorkerPool[T]) Add(task Task[T]) {
	if p.ctx.Err() != nil {
		return
	}
	p.tasks <- task
}

func (p *WorkerPool[T]) Wait() {
	close(p.tasks)
	p.wg.Wait()
	close(p.results)
}

func (p *WorkerPool[T]) Result() Result[T] {
	if p.ctx.Err() != nil {
		return Result[T]{Error: p.ctx.Err()}
	}
	return <-p.results
}

func (p *WorkerPool[T]) IsCancelled() bool {
	if errors.Is(p.ctx.Err(), context.Canceled) {
		return true
	}
	return false
}

func (p *WorkerPool[T]) Task() <-chan Task[T] {
	return p.tasks
}

func (p *WorkerPool[T]) worker() {
	defer p.wg.Done()

	for task := range p.tasks {
		if p.ctx.Err() != nil {
			return
		}
		p.results <- task.Exec()
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	pool := NewWorkerPool[int](ctx, 4)
	pool.Run()
	for i := 1; i <= 15; i++ {
		i := i
		task := NewTask[int](func() Result[int] {
			fmt.Printf("gorutines count: %v\n", runtime.NumGoroutine())
			return Result[int]{Ok: i * i, Error: nil}
		})
		pool.Add(task)

		result := pool.Result()
		if pool.IsCancelled() {
			continue
		}
		fmt.Printf("Получен результат обработки задачи %d\n", result.Ok)
		if result.Ok == 5 {
			cancel()
		}
	}
	pool.Wait()
	time.Sleep(15 * time.Second)
	fmt.Printf("end work main\ngorutines count: %v\n", runtime.NumGoroutine())
}
