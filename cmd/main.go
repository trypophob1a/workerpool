package main

import (
	"sync"
	"sync/atomic"
	"time"
)

type DynamicBuffer struct {
	lock          *sync.RWMutex
	maxBufferSize int
	buffer        chan Task
	cap           int64
}

func NewDynamicBuffer(initialSize, maxBufferSize int) *DynamicBuffer {
	return &DynamicBuffer{
		lock:          &sync.RWMutex{},
		maxBufferSize: maxBufferSize,
		buffer:        make(chan Task, initialSize),
		cap:           int64(initialSize),
	}
}

func (db *DynamicBuffer) Add(task Task) {
	db.lock.Lock()
	canGrow := len(db.buffer) == cap(db.buffer) && cap(db.buffer) < db.maxBufferSize
	if canGrow {
		db.buffer = db.grow()
	}
	db.lock.Unlock()
	db.buffer <- task
}

func (db *DynamicBuffer) Cap() int {
	db.lock.RLock()
	defer db.lock.RUnlock()
	return cap(db.buffer)
}

func (db *DynamicBuffer) GetCap() int64 {
	return atomic.LoadInt64(&db.cap)
}

func (db *DynamicBuffer) Len() int {
	db.lock.RLock()
	defer db.lock.RUnlock()
	return len(db.buffer)
}

func (db *DynamicBuffer) grow() chan Task {
	newCap := cap(db.buffer) * 2
	if newCap > db.maxBufferSize {
		newCap = db.maxBufferSize
	}

	db.cap = int64(newCap)
	newBuffer := make(chan Task, newCap)
	close(db.buffer)
	for task := range db.buffer {
		newBuffer <- task
	}
	return newBuffer
}

type Task func()

type WorkerPool struct {
	maxWorkers    int
	waitingTasks  *DynamicBuffer
	wg            *sync.WaitGroup
	Len           int64
	maxBufferSize int64
	spaceCond     sync.Cond
	semaphore     chan struct{}
}

func NewPool(maxWorkers, bufferSize int) *WorkerPool {
	return &WorkerPool{
		maxWorkers:    maxWorkers,
		wg:            &sync.WaitGroup{},
		waitingTasks:  NewDynamicBuffer(2, bufferSize),
		maxBufferSize: int64(bufferSize),
		spaceCond:     *sync.NewCond(&sync.Mutex{}),
		semaphore:     make(chan struct{}, bufferSize),
	}
}

func (p *WorkerPool) Run() {
	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *WorkerPool) Submit(task Task) {
	//	p.spaceCond.L.Lock()
	//	for p.IsFull() {
	//	p.spaceCond.Wait() // ждем оповещения
	//		println("wait")
	//	}
	if p.IsFull() {
		<-p.semaphore
	}
	p.waitingTasks.Add(task)
	atomic.AddInt64(&p.Len, 1)
	// p.spaceCond.L.Unlock()
}

func (p *WorkerPool) IsFull() bool {
	return atomic.LoadInt64(&p.Len) == p.maxBufferSize
}

func (p *WorkerPool) hasSpace() bool {
	return p.waitingTasks.GetCap() == p.maxBufferSize && atomic.LoadInt64(&p.Len)+1 >= p.maxBufferSize
}

func (p *WorkerPool) worker() {
	defer p.wg.Done()
	for {
		task, ok := <-p.waitingTasks.buffer
		if !ok {
			return
		}
		task()
		atomic.AddInt64(&p.Len, -1)
		if p.hasSpace() {
			//	p.spaceCond.Broadcast()
			println("has space")
			p.semaphore <- struct{}{}
		}
	}
}

func (p *WorkerPool) Wait() {
	close(p.waitingTasks.buffer)
	p.wg.Wait()
}

func main() {
	pool := NewPool(2, 16)

	pool.Run()
	for i := 1; i <= 5; i++ {
		i := i
		pool.Submit(func() {
			println(pool.waitingTasks.Len(), pool.waitingTasks.Cap(), pool.waitingTasks.GetCap())
			time.Sleep(2 * time.Second)
			println("slow", i)
			// slow <- i
		})

		pool.Submit(func() {
			println(pool.waitingTasks.Len(), pool.waitingTasks.Cap(), pool.waitingTasks.GetCap())
			println("fast", i)
		})
		pool.Submit(func() {
			println(pool.waitingTasks.Len(), pool.waitingTasks.Cap(), pool.waitingTasks.GetCap())
			println("fast2", i)
		})
		pool.Submit(func() {
			time.Sleep(2 * time.Second)
			println(pool.waitingTasks.Len(), pool.waitingTasks.Cap(), pool.waitingTasks.GetCap())
			println("slow2", i)
		})
	}
	pool.Wait()
}
