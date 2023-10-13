package workerpool

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestDynamicBuffer_Add(t *testing.T) {
	t.Run("race check add", func(t *testing.T) {
		db := NewDynamicBuffer(10, 100)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				db.Add(NewWorker(context.Background(), func(ctx context.Context) {}))
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				db.Add(NewWorker(context.Background(), func(ctx context.Context) {}))
			}
		}()

		wg.Wait()

		if db.Cap() != 100 {
			t.Errorf("expected cap 100, got %d", db.Cap())
		}
	})
	t.Run("race check read buffer after add", func(t *testing.T) {
		db := NewDynamicBuffer(2, 2)
		var wg sync.WaitGroup
		for i := 0; i < 1; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					w, ok := <-db.buffer
					if !ok {
						return
					}
					w.Task(w.Ctx)
				}
			}()
		}
		nums := make(chan int)
		sum := 0
		for i := 1; i < 4; i++ {
			i := i
			db.Add(NewWorker(context.Background(), func(ctx context.Context) {
				nums <- i
			}))
			sum += <-nums
		}
		close(db.buffer)
		wg.Wait()
		if sum != 6 {
			t.Errorf("expected 6, got %d", sum)
		}
		if db.Cap() != 2 {
			t.Errorf("expected cap 2, got %d", db.Cap())
		}
	})
	t.Run("buffer is full", func(t *testing.T) {
		db := NewDynamicBuffer(1, 1)
		db.Add(NewWorker(context.Background(), func(ctx context.Context) {}))
		if ok := db.Add(NewWorker(context.Background(), func(ctx context.Context) {})); ok {
			t.Errorf("expected false, got true")
		}
	})
	t.Run("buffer is empty", func(t *testing.T) {
		db := NewDynamicBuffer(1, 1)
		if ok := db.Add(NewWorker(context.Background(), func(ctx context.Context) {})); !ok {
			t.Errorf("expected true, got false")
		}
	})
	t.Run("the buffer gets the value", func(t *testing.T) {
		db := NewDynamicBuffer(2, 4)
		sum := 0
		db.Add(NewWorker(context.Background(), func(ctx context.Context) {
			sum += 1
		}))
		worker := <-db.buffer
		worker.Task(worker.Ctx)
		db.Add(NewWorker(context.Background(), func(ctx context.Context) {
			sum += 1
		}))
		worker = <-db.buffer
		worker.Task(worker.Ctx)
		if sum != 2 {
			t.Errorf("expected 2, got %d", sum)
		}
	})
}

func TestDynamicBuffer_Cap(t *testing.T) {
	t.Run("empty buffer", func(t *testing.T) {
		db := &DynamicBuffer{buffer: make(chan *Worker, 2), cap: 2, maxBufferSize: 4}
		db.buffer <- NewWorker(context.Background(), func(ctx context.Context) {})
		if db.Cap() != 2 {
			t.Errorf("expected cap 2, got %d", db.Cap())
		}
	})
	t.Run("Race check", func(t *testing.T) {
		db := NewDynamicBuffer(2, 2)
		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			for i := 0; i < 2; i++ {
				db.Add(NewWorker(context.Background(), func(ctx context.Context) {}))
			}
		}()

		go func() {
			defer wg.Done()
			if db.Cap() != 2 {
				t.Errorf("expected cap 2, got %d", db.Cap())
			}
		}()
		wg.Wait()
		if db.Cap() != 2 {
			t.Errorf("expected cap 2, got %d", db.Cap())
		}
	})

	t.Run("after grow", func(t *testing.T) {
		db := NewDynamicBuffer(2, 4)
		db.Add(NewWorker(context.Background(), func(ctx context.Context) {}))
		if db.Cap() != 2 {
			t.Errorf("expected cap 2, got %d", db.Cap())
		}
		db.Add(NewWorker(context.Background(), func(ctx context.Context) {}))
		db.Add(NewWorker(context.Background(), func(ctx context.Context) {}))
		if db.Cap() != 4 {
			t.Errorf("expected cap 4, got %d", db.Cap())
		}
	})
}

func TestDynamicBuffer_grow(t *testing.T) {
	t.Run("empty buffer", func(t *testing.T) {
		db := NewDynamicBuffer(1, 2)
		if ok := db.Add(NewWorker(context.Background(), func(ctx context.Context) {})); !ok {
			t.Errorf("expected true, got false")
		}
		ok := db.Add(NewWorker(context.Background(), func(ctx context.Context) {}))
		if !ok {
			t.Errorf("expected true, got false")
		}
		if db.Cap() != 2 {
			t.Errorf("expected cap 2, got %d", db.Cap())
		}
	})
}

func TestNewDynamicBuffer(t *testing.T) {
	db := NewDynamicBuffer(2, 4)

	expected := &DynamicBuffer{
		lock:          &sync.RWMutex{},
		maxBufferSize: 4,
		buffer:        make(chan *Worker, 2),
		cap:           2,
	}

	if len(expected.buffer) != len(db.buffer) {
		t.Errorf("expected %d, got %d", len(expected.buffer), len(db.buffer))
	}

	if expected.maxBufferSize != db.maxBufferSize {
		t.Errorf("expected %d, got %d", expected.maxBufferSize, db.maxBufferSize)
	}
	if expected.cap != db.cap {
		t.Errorf("expected %d, got %d", expected.cap, db.cap)
	}

	if T := fmt.Sprintf("%v", reflect.TypeOf(expected.buffer)); T != "chan *workerpool.Worker" {
		t.Errorf("expected %s, got %s", "chan *workerpool.Worker", T)
	}
}
