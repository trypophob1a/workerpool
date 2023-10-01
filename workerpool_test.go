package workerpool

import (
	"context"
	"fmt"
	"runtime"
	"testing"
)

func examp(a int) (int, error) {
	if a < 0 {
		return 0, fmt.Errorf("error a is %d < 0", a)
	}
	return a, nil
}

func TestNewTask(t *testing.T) {
	numJobs := 10
	numWorkers := 1
	p := NewWorkerPool[int](numWorkers)
	p.Run()
	for i := 1; i < numJobs; i++ {
		index := i
		p.Add(NewTask[int](context.Background(), func() Result[int] {
			val, err := examp(index)
			return Result[int]{
				Ok:    val,
				Error: err,
			}
		}))
	}
	p.Close()
	fmt.Printf("%d\n", runtime.NumGoroutine())
	p.Wait(func(res Result[int]) {
		fmt.Printf("from wait: val: %d, err: %v\n", res.Ok, res.Error)
	})
}
