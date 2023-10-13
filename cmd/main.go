package main

import (
	"context"
	"time"

	"github.com/trypophob1a/workerpool"
)

func main() {
	pool := workerpool.NewPool(2, 20)

	pool.Run()
	for i := 1; i <= 5; i++ {
		i := i
		pool.Submit(workerpool.NewWorker(context.Background(), func(ctx context.Context) {
			//println(pool.len(), pool.workers.Cap())
			time.Sleep(2 * time.Second)
			if ctx.Err() != nil {
				println("slow1 canceled!!!!!!!")
				return
			}
		}))
		pool.Submit(workerpool.NewWorker(context.Background(), func(ctx context.Context) {
			//println(pool.len(), pool.workers.Cap())
			println("fast1", i)
		}))
		pool.Submit(workerpool.NewWorker(context.Background(), func(ctx context.Context) {
			//println(pool.len(), pool.workers.Cap())
			println("fast2", i)
		}))
		pool.Submit(workerpool.NewWorker(context.Background(), func(ctx context.Context) {
			//println(pool.len(), pool.workers.Cap())
			time.Sleep(2 * time.Second)
			println("slow2", i)
		}))
	}
	pool.Wait()
	println("end")
}
