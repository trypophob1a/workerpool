package queue

import (
	"fmt"
	"strings"
	"sync"
)

type node[T any] struct {
	value T
	next  *node[T]
	prev  *node[T]
}

type Queue[T any] struct {
	mu   *sync.RWMutex
	head *node[T]
	tail *node[T]
	size int
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		mu:   &sync.RWMutex{},
		head: nil,
		tail: nil,
		size: 0,
	}
}

func (q *Queue[T]) Enqueue(value T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.size++
	if q.size == 1 {
		q.head = &node[T]{
			value: value,
			next:  nil,
			prev:  nil,
		}
		q.tail = q.head
	} else {
		q.tail.next = &node[T]{
			value: value,
			next:  nil,
			prev:  q.tail,
		}
		q.tail = q.tail.next
	}
}

// Dequeue extracts and removes an element from the front of the queue.
// If the queue is empty, it returns the default value and false.
// Otherwise, it returns the value from the queue's head and true.
//
// Example:
//
//	// Create a new queue
//	q := NewQueue[int]()
//
//	// Add elements to the queue
//	q.Enqueue(1)
//	q.Enqueue(2)
//	q.Enqueue(3)
//
//	// Display the current state of the queue
//	fmt.Printf("Current queue: %v\n", q.String()) // Output: Current queue: [1 2 3]
//
//	// Dequeue an element from the head of the queue
//	val, ok := q.Dequeue() // in value 1 ok true
//	fmt.Printf("Obtaining %s in value %v ok %v\n", q.String(), val, ok) // Output: Obtaining [2 3]
func (q *Queue[T]) Dequeue() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.size == 0 {
		var t T
		return t, false
	}
	res := q.head.value
	q.size--
	if q.size == 0 {
		q.head = nil
		q.tail = nil
	} else {
		q.head = q.head.next
		q.head.prev = nil
	}
	return res, true
}

func (q *Queue[T]) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.size
}

func (q *Queue[T]) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.size == 0
}

func (q *Queue[T]) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.size = 0
	q.head = nil
	q.tail = nil
}

func (q *Queue[T]) Peek() (T, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.size == 0 {
		var t T
		return t, false
	}
	return q.head.value, true
}

func (q *Queue[T]) ForEach(callback func(n *node[T])) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	for head := q.head; head != nil; head = head.next {
		callback(head)
	}
}

func (q *Queue[T]) String() string {
	var sb strings.Builder
	sb.WriteRune('[')
	q.ForEach(func(n *node[T]) {
		if n == q.head {
			sb.WriteString(fmt.Sprintf("%v", n.value))
		} else {
			sb.WriteString(fmt.Sprintf(" %v", n.value))
		}
	})
	sb.WriteRune(']')
	return sb.String()
}
