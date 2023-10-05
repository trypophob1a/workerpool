package main

import (
	"context"
	"testing"
)

func BenchmarkWithoutPool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < b.N; i++ {
			for i := 1; i <= 15; i++ {
				i := i
				task := func() Result[int] {
					//fmt.Printf("gorutines count: %v\n", runtime.NumGoroutine())
					return Result[int]{Ok: i * i, Error: nil}
				}

				result := task()
				result.Ok += result.Ok
			}
		}
	}
	b.StopTimer()
}
func BenchmarkPool(b *testing.B) {
	b.ResetTimer()
	ctx, _ := context.WithCancel(context.Background())
	for i := 0; i < b.N; i++ {
		pool := NewWorkerPool[int](ctx, 2)
		pool.Run()
		for i := 1; i <= 15; i++ {
			i := i
			task := NewTask[int](func() Result[int] {
				//fmt.Printf("gorutines count: %v\n", runtime.NumGoroutine())
				return Result[int]{Ok: i * i, Error: nil}
			})
			pool.Add(task)

			result := pool.Result()
			result.Ok += result.Ok
		}
		pool.Wait()
	}
	b.StopTimer()
}

type fn[T any] func() Result[T]
