package workerpool

import (
	"context"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	pool := NewPool(2, 20)
	expect := &WorkerPool{
		maxWorkers:    2,
		wg:            &sync.WaitGroup{},
		workers:       NewDynamicBuffer(2, 20),
		maxBufferSize: 20,
		semaphore:     make(chan struct{}, 20),
	}

	if len(expect.workers.buffer) != len(pool.workers.buffer) {
		t.Errorf("expected %d, got %d",
			len(expect.workers.buffer),
			len(pool.workers.buffer),
		)
	}

	if cap(expect.workers.buffer) != cap(pool.workers.buffer) {
		t.Errorf("expected %d, got %d",
			cap(expect.workers.buffer),
			cap(pool.workers.buffer),
		)
	}

	if expect.maxBufferSize != pool.maxBufferSize {
		t.Errorf("expected %d, got %d",
			expect.maxBufferSize,
			pool.maxBufferSize,
		)
	}

	if expect.maxWorkers != pool.maxWorkers {
		t.Errorf("expected %d, got %d",
			expect.maxWorkers,
			pool.maxWorkers,
		)
	}
}

func TestNewWorker(t *testing.T) {
	worker := NewWorker(context.Background(), func(ctx context.Context) {})
	expected := &Worker{
		Ctx:     context.Background(),
		TimeOut: nil,
		Task:    func(ctx context.Context) {},
	}
	workerTask := reflect.TypeOf(worker.Task).String()
	expectedTask := reflect.TypeOf(expected.Task).String()
	if workerTask != expectedTask {
		t.Errorf("expected %s, got %s", expectedTask, workerTask)
	}
}

func TestWorkerPool_IsFull(t *testing.T) {
	t.Run("full", func(t *testing.T) {
		pool := NewPool(2, 2)
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {}))
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {}))
		if !pool.IsFull() {
			t.Errorf("expected true, got false")
		}
	})
	t.Run("Race check", func(t *testing.T) {
		pool := NewPool(2, 10)

		var wg sync.WaitGroup
		wg.Add(10)

		for i := 0; i < 10; i++ {
			go func() {
				defer wg.Done()
				if pool.IsFull() {
					return
				}
				pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {}))
			}()
		}

		wg.Wait()

		if pool.Len != 10 {
			t.Errorf("expected pool length 10, got %d", pool.len())
		}
	})
	t.Run("Not full", func(t *testing.T) {
		pool := NewPool(2, 3)
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {}))
		if pool.IsFull() {
			t.Errorf("expected false, got true")
		}
	})
}

func TestWorkerPool_Run(t *testing.T) {
	t.Run("1 worker", func(t *testing.T) {
		const workerCount = 1
		pool := NewPool(workerCount, 20)
		numGoroutineBeforeRun := runtime.NumGoroutine()
		pool.Run()
		numGoroutineAfterRun := runtime.NumGoroutine()
		wantGoroutines := numGoroutineAfterRun - numGoroutineBeforeRun
		if wantGoroutines != workerCount {
			t.Errorf("expected %d, got %d", wantGoroutines, workerCount)
		}
		pool.Wait()
	})
	t.Run("2 worker", func(t *testing.T) {
		const workerCount = 1
		pool := NewPool(workerCount, 20)
		numGoroutineBeforeRun := runtime.NumGoroutine()
		pool.Run()
		numGoroutineAfterRun := runtime.NumGoroutine()
		wantGoroutines := numGoroutineAfterRun - numGoroutineBeforeRun
		if wantGoroutines != workerCount {
			t.Errorf("expected %d, got %d", wantGoroutines, workerCount)
		}
		pool.Wait()
	})
	t.Run("3 workers", func(t *testing.T) {
		const workerCount = 3
		pool := NewPool(workerCount, 20)
		numGoroutineBeforeRun := runtime.NumGoroutine()
		pool.Run()
		numGoroutineAfterRun := runtime.NumGoroutine()
		wantGoroutines := numGoroutineAfterRun - numGoroutineBeforeRun
		if wantGoroutines != workerCount {
			t.Errorf("expected %d, got %d", wantGoroutines, workerCount)
		}
		pool.Wait()
	})

}

func TestWorkerPool_Submit(t *testing.T) {
	t.Run("two elements submitted", func(t *testing.T) {
		pool := NewPool(2, 2)
		nums := make(chan int, 2)
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
			nums <- 1
		}))
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
			nums <- 2
		}))
		if pool.len() != 2 {
			t.Errorf("expected 2, got %d", pool.len())
		}
		worker := <-pool.workers.buffer
		worker.Task(context.Background())
		worker = <-pool.workers.buffer
		worker.Task(context.Background())
		sum := 0
		for i := 0; i < 2; i++ {
			sum += <-nums
		}
		close(pool.workers.buffer)
		close(nums)
		if sum != 3 {
			t.Errorf("expected 3, got %d", sum)
		}
	})
	t.Run("race check", func(t *testing.T) {
		pool := NewPool(2, 2)
		nums := make(chan int, 2)
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
				nums <- 1
			}))
		}()
		go func() {
			defer wg.Done()
			pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
				nums <- 2
			}))
		}()
		wg.Wait()
		if pool.len() != 2 {
			t.Errorf("expected 2, got %d", pool.len())
		}
		worker := <-pool.workers.buffer
		worker.Task(context.Background())
		worker = <-pool.workers.buffer
		worker.Task(context.Background())
		sum := 0
		for i := 0; i < 2; i++ {
			sum += <-nums
		}
		close(pool.workers.buffer)
		close(nums)
		if sum != 3 {
			t.Errorf("expected 3, got %d", sum)
		}
	})

	t.Run("is full", func(t *testing.T) {
		pool := NewPool(2, 2)
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {}))
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {}))
		if pool.len() != 2 {
			t.Errorf("expected false, got true")
		}
		pool.semaphore <- struct{}{}
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {}))

		if pool.len() != 2 {
			t.Errorf("expected false, got true")
		}
	})
	t.Run("with timeout", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			pool := NewPool(2, 2)
			pool.Run()
			nums := make(chan int, 3)
			for i := 1; i <= 3; i++ {
				i := i
				pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
					if i%2 == 0 {
						time.Sleep(100 * time.Millisecond)
					}
					if ctx.Err() != nil {
						return
					}
					nums <- i
				}, time.Millisecond*50))
			}
			pool.Wait()
			close(nums)
			sum := 0
			for num := range nums {
				sum += num
			}
			if sum != 4 {
				t.Errorf("expected 4, got %d", sum)
			}
		}
	})
}

func TestWorkerPool_Wait(t *testing.T) {
	pool := NewPool(2, 2)
	pool.Run()
	nums := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		i := i
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
			nums <- i
		}))
	}
	pool.Wait()
	close(nums)
	sum := 0
	for num := range nums {
		sum += num
	}
	if sum != 15 {
		t.Errorf("expected 15, got %d", sum)
	}
}

func TestWorkerPool_hasSpace(t *testing.T) {
	t.Run("has space", func(t *testing.T) {
		pool := NewPool(2, 2)
		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
			}))
		}()

		go func() {
			defer wg.Done()
			pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
			}))
		}()

		wg.Wait() // Ожидаем завершения обеих горутин
		<-pool.workers.buffer
		pool.Len--
		if !pool.hasSpace() {
			t.Errorf("Expected true, got false")
		}
	})
}

func TestWorkerPool_len(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		pool := NewPool(2, 2)
		if pool.len() != 0 {
			t.Errorf("expected 0, got %d", pool.len())
		}
	})
	t.Run("add two elements", func(t *testing.T) {
		pool := NewPool(2, 2)
		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
			}))
		}()

		go func() {
			defer wg.Done()
			pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
			}))
		}()
	})
}

func TestWorkerPool_worker(t *testing.T) {
	t.Run("two elements submitted", func(t *testing.T) {
		pool := NewPool(2, 2)
		nums := make(chan int, 2)
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
			nums <- 1
		}))
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
			nums <- 2
		}))
		if pool.len() != 2 {
			t.Errorf("expected 2, got %d", pool.len())
		}
		worker := <-pool.workers.buffer
		worker.Task(context.Background())
		worker = <-pool.workers.buffer
		worker.Task(context.Background())
		sum := 0
		for i := 0; i < 2; i++ {
			sum += <-nums
		}
		close(pool.workers.buffer)
		close(nums)
		if sum != 3 {
			t.Errorf("expected 3, got %d", sum)
		}
	})
	t.Run("race check", func(t *testing.T) {
		pool := NewPool(2, 2)
		nums := make(chan int, 2)
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
				nums <- 1
			}))
		}()
		go func() {
			defer wg.Done()
			pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {
				nums <- 2
			}))
		}()
		wg.Wait()
		if pool.len() != 2 {
			t.Errorf("expected 2, got %d", pool.len())
		}
		worker := <-pool.workers.buffer
		worker.Task(context.Background())
		worker = <-pool.workers.buffer
		worker.Task(context.Background())
		sum := 0
		for i := 0; i < 2; i++ {
			sum += <-nums
		}
		close(pool.workers.buffer)
		close(nums)
		if sum != 3 {
			t.Errorf("expected 3, got %d", sum)
		}
	})

	t.Run("is full", func(t *testing.T) {
		pool := NewPool(2, 2)
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {}))
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {}))
		if pool.len() != 2 {
			t.Errorf("expected false, got true")
		}
		pool.semaphore <- struct{}{}
		pool.Submit(NewWorker(context.Background(), func(ctx context.Context) {}))

		if pool.len() != 2 {
			t.Errorf("expected false, got true")
		}
	})
}

func TestWorker_Kill(t *testing.T) {
	i := 0
	worker := NewWorker(context.Background(), func(ctx context.Context) {
		if ctx.Err() != nil {
			return
		}
		i++
	}, time.Millisecond*1)
	worker.Kill()
	worker.Task(worker.Ctx)
	if i != 0 {
		t.Errorf("expected 0, got %d", i)
	}
}
