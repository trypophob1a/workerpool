package queue

import (
	"fmt"
	"strings"
	"sync"
)

type Node[T any] struct {
	Value T
	next  *Node[T]
	prev  *Node[T]
}

type Queue[T any] struct {
	mu   *sync.RWMutex
	head *Node[T]
	tail *Node[T]
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
		q.head = &Node[T]{
			Value: value,
			next:  nil,
			prev:  nil,
		}
		q.tail = q.head
	} else {
		q.tail.next = &Node[T]{
			Value: value,
			next:  nil,
			prev:  q.tail,
		}
		q.tail = q.tail.next
	}
}

// Dequeue extracts and removes an element from the front of the queue.
// If the queue is empty, it returns the default Value and false.
// Otherwise, it returns the Value from the queue's head and true.
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
//	val, ok := q.Dequeue() // in Value 1 ok true
//	fmt.Printf("Obtaining %s in Value %v ok %v\n", q.String(), val, ok) // Output: Obtaining [2 3]
func (q *Queue[T]) Dequeue() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.size == 0 {
		var t T
		return t, false
	}
	res := q.head.Value
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
	return q.head.Value, true
}

func (q *Queue[T]) ForEach(callback func(n *Node[T])) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	for head := q.head; head != nil; head = head.next {
		callback(head)
	}
}

func (q *Queue[T]) String() string {
	var sb strings.Builder
	sb.WriteRune('[')
	q.ForEach(func(n *Node[T]) {
		if n == q.head {
			sb.WriteString(fmt.Sprintf("%v", n.Value))
		} else {
			sb.WriteString(fmt.Sprintf(" %v", n.Value))
		}
	})
	sb.WriteRune(']')
	return sb.String()
}
