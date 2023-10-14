package main

import (
	"context"

	"github.com/trypophob1a/workerpool"
)

func main() {
	pool := workerpool.NewPool(2, 1000)
	nums := make(chan int)
	pool.Run()
	sum := 0

	for i := 1; i <= 50000; i++ {
		i := i
		//pool.Submit(workerpool.NewWorker(context.Background(), func(ctx context.Context) {
		//	//println(pool.len(), pool.workers.Cap())
		//	time.Sleep(2 * time.Second)
		//	if ctx.Err() != nil {
		//		println("slow1 canceled!!!!!!!")
		//		return
		//	}
		//}, time.Millisecond*50))
		pool.Submit(workerpool.NewWorker(context.Background(), func(ctx context.Context) {
			//println(pool.len(), pool.workers.Cap())
			println("fast1", i)
			nums <- i
		}))
		pool.Submit(workerpool.NewWorker(context.Background(), func(ctx context.Context) {
			//println(pool.len(), pool.workers.Cap())
			println("fast2", i)
		}))
		pool.Submit(workerpool.NewWorker(context.Background(), func(ctx context.Context) {
			//println(pool.len(), pool.workers.Cap())
			//time.Sleep(2 * time.Second)
			println("slow2", i)
		}))
		sum += <-nums
	}
	pool.Wait()
	close(nums)
	println(sum)
	println("end")
}
