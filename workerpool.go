package workerpool

import (
	"sync"
)

type Result[T any] struct {
	Ok    T
	Error error
}

type Task[T any] struct {
	Process func() Result[T]
}

func NewTask[T any](process func() Result[T]) *Task[T] {
	return &Task[T]{Process: process}
}

func NewResult[T any]() Result[T] {
	return Result[T]{}
}

type WorkerPool[T any] struct {
	workersCount int
	wg           *sync.WaitGroup
	results      chan Result[T]
	tasks        chan *Task[T]
}

func NewWorkerPool[T any](workersCount int, tasks chan *Task[T]) *WorkerPool[T] {
	pool := &WorkerPool[T]{
		workersCount: workersCount,
		results:      make(chan Result[T], workersCount),
		tasks:        tasks,
		wg:           &sync.WaitGroup{},
	}
	pool.Run()
	return pool
}

func (w *WorkerPool[T]) Add(task *Task[T]) {
	w.tasks <- task
}
func (w *WorkerPool[T]) worker() {
	for task := range w.tasks {
		w.results <- task.Process()
	}
}
func (w *WorkerPool[T]) Close() {
	close(w.tasks)
}
func (w *WorkerPool[T]) Run() {
	// Создание воркеров
	for i := 0; i < w.workersCount; i++ {
		w.wg.Add(1)
		go func(id int) {
			defer w.wg.Done()
			w.worker()
		}(i)
	}

}
func (w *WorkerPool[T]) Wait(callback func(Result[T])) {
	close(w.tasks)
	go func() {
		w.wg.Wait()
		close(w.results)
	}()
	for result := range w.results {
		callback(result)
	}
}
