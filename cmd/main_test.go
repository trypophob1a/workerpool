package main

import (
	"context"
	"testing"

	"github.com/gammazero/workerpool"
)

func BenchmarkGammaZeroPool(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pool := workerpool.New(2)
		result := make(chan int)
		result2 := make(chan int)
		for i := 1; i <= 15; i++ {
			i := i

			pool.Submit(func() {
				result <- i * i
			})
			pool.Submit(func() {
				result2 <- i * i
			})
			res1 := <-result
			res1 += res1
			res2 := <-result2
			res2 += res2
		}
		pool.StopWait()
	}
	b.StopTimer()
}
func BenchmarkMyPool(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pool := NewPool(2, 16)
		pool.Run()
		result := make(chan int)
		result2 := make(chan int)
		for i := 1; i <= 15; i++ {
			i := i
			task := NewWorker(context.Background(), func(ctx context.Context) {
				result <- i * i
			})
			pool.Submit(task)
			task2 := NewWorker(context.Background(), func(ctx context.Context) {
				result2 <- i * i
			})
			pool.Submit(task2)
			res := <-result
			res += res
			res2 := <-result2
			res2 += res2
		}
		pool.Wait()
	}
	b.StopTimer()
}
