package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
)

type result struct {
	Ok  int
	Err error
}

type que struct {
	mu *sync.Mutex
	q  []task
}

func newQue() *que {
	return &que{
		mu: &sync.Mutex{},
		q:  make([]task, 0),
	}
}

func (q *que) push(t task) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.q = append(q.q, t)
}

func (q *que) pop() (task, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.q) == 0 {
		return nil, false
	}
	t := q.q[0]
	q.q = q.q[1:]
	return t, true
}

type task func() result

type pool struct {
	wg      *sync.WaitGroup
	tasks   *que
	results chan result
	wc      int
	ctx     context.Context
}

func newPool(ctx context.Context, wc int) *pool {
	return &pool{
		tasks:   newQue(),
		wg:      &sync.WaitGroup{},
		results: make(chan result),
		wc:      wc,
		ctx:     ctx,
	}
}

func (p *pool) push(t task) {
	p.tasks.push(t)
}
func (p *pool) worker() error {
	if err := p.ctx.Err(); err != nil {
		println("cancel")
		return err
	}

	t, ok := p.tasks.pop()
	if !ok {
		return fmt.Errorf("queue is empty")
	}
	p.results <- t()
	return nil
}

func (p *pool) run() {
	for i := 0; i < p.wc; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				err := p.worker()
				if err != nil {
					return
				}
			}
		}()
	}
}

func (p *pool) wait(callback func(result result)) {
	go func() {
		p.wg.Wait()
		close(p.results)
	}()
	for r := range p.results {
		callback(r)
	}
}
func main() {
	ctx, cancel := context.WithCancel(context.Background())

	p := newPool(ctx, 2)
	p.run()
	for i := 0; i <= 35; i++ {
		index := i
		p.push(func() result {
			r := result{}
			if index%2 == 0 {
				r.Ok = index
			} else {
				r.Err = fmt.Errorf("error: %d", index)
			}
			return r
		})
	}

	p.wait(func(r result) {
		fmt.Printf("ok: %v, err: %v\n", r.Ok, r.Err)
		if r.Ok == 18 {
			cancel()
		}
	})
	fmt.Printf("end last gorutines: %d\n", runtime.NumGoroutine())
}
