package main

import (
	"context"
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
	type fields struct {
		lock          *sync.RWMutex
		maxBufferSize int
		buffer        chan *Worker
		cap           int64
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &DynamicBuffer{
				lock:          tt.fields.lock,
				maxBufferSize: tt.fields.maxBufferSize,
				buffer:        tt.fields.buffer,
				cap:           tt.fields.cap,
			}
			if got := db.Cap(); got != tt.want {
				t.Errorf("Cap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDynamicBuffer_grow(t *testing.T) {
	type fields struct {
		lock          *sync.RWMutex
		maxBufferSize int
		buffer        chan *Worker
		cap           int64
	}
	tests := []struct {
		name   string
		fields fields
		want   chan *Worker
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &DynamicBuffer{
				lock:          tt.fields.lock,
				maxBufferSize: tt.fields.maxBufferSize,
				buffer:        tt.fields.buffer,
				cap:           tt.fields.cap,
			}
			if got := db.grow(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("grow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDynamicBuffer(t *testing.T) {
	type args struct {
		initialSize   int
		maxBufferSize int
	}
	tests := []struct {
		name string
		args args
		want *DynamicBuffer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDynamicBuffer(tt.args.initialSize, tt.args.maxBufferSize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDynamicBuffer() = %v, want %v", got, tt.want)
			}
		})
	}
}
