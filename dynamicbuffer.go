package workerpool

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

// Buffer returns a channel of *Worker from the DynamicBuffer.
//
// This function does not take any parameters.
// It returns a chan *Worker.
func (db *DynamicBuffer) Buffer() chan *Worker {
	db.lock.RLock()
	defer db.lock.RUnlock()
	return db.buffer
}

// Add appends a task to the DynamicBuffer.
//
// This method adds the specified task to the buffer, provided that the buffer is not
// already at its maximum capacity. If the buffer is full, it returns false to indicate
// that the task could not be added.
//
// If the buffer is at its maximum capacity and can be grown further, this method will
// automatically double the capacity of the buffer using the `grow` method and transfer
// existing tasks to the new buffer.
//
// Parameters:
//
//	task (*Worker): The task to be added to the buffer.
//
// Returns:
//
//	bool: True if the task was successfully added, false if the buffer is at its maximum
//	      capacity and cannot be grown further.
func (db *DynamicBuffer) Add(task *Worker) bool {
	db.lock.Lock()
	defer db.lock.Unlock()

	if len(db.buffer) == db.maxBufferSize {
		return false
	}

	canGrow := len(db.buffer) == cap(db.buffer) && cap(db.buffer) < db.maxBufferSize

	if canGrow {
		db.buffer = db.grow()
	}

	db.buffer <- task
	return true
}

// Cap returns the current capacity of the DynamicBuffer.
func (db *DynamicBuffer) Cap() int64 {
	return atomic.LoadInt64(&db.cap)
}

// grow dynamically increases the size of the buffer.
//
// It doubles the capacity of the buffer and creates a new channel with the
// updated capacity. If the new capacity exceeds the maximum buffer size, it
// is capped to the maximum buffer size. The existing tasks in the buffer are
// transferred to the new buffer. The new buffer is returned.
//
// Returns:
//
//	chan *Worker: The new buffer with increased capacity.
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
