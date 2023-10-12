package main

import (
	"sync"
	"sync/atomic"
)

type DynamicBuffer struct {
	lock          *sync.RWMutex
	maxBufferSize int
	buffer        chan *Worker
	cap           int64
}

func NewDynamicBuffer(initialSize, maxBufferSize int) *DynamicBuffer {
	return &DynamicBuffer{
		lock:          &sync.RWMutex{},
		maxBufferSize: maxBufferSize,
		buffer:        make(chan *Worker, initialSize),
		cap:           int64(initialSize),
	}
}

func (db *DynamicBuffer) Add(task *Worker) bool {
	db.lock.RLock()
	if len(db.buffer) == db.maxBufferSize {
		return false
	}
	db.lock.RUnlock()
	db.lock.Lock()
	defer db.lock.Unlock()
	canGrow := len(db.buffer) == cap(db.buffer) && cap(db.buffer) < db.maxBufferSize

	if canGrow {
		db.buffer = db.grow()
	}

	db.buffer <- task
	return true
}

func (db *DynamicBuffer) Cap() int64 {
	return atomic.LoadInt64(&db.cap)
}

func (db *DynamicBuffer) grow() chan *Worker {
	newCap := cap(db.buffer) * 2
	if newCap > db.maxBufferSize {
		newCap = db.maxBufferSize
	}

	newBuffer := make(chan *Worker, newCap)
	db.cap = int64(newCap)
	close(db.buffer)
	for task := range db.buffer {
		newBuffer <- task
	}

	return newBuffer
}
