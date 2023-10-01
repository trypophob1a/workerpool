package queue

import (
	"reflect"
	"sync"
	"testing"
)

func TestNewQueue(t *testing.T) {
	type testCase[T any] struct {
		name string
		want *Queue[T]
	}
	tests := []testCase[int]{
		{"NewQueue", &Queue[int]{&sync.RWMutex{}, nil, nil, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewQueue[int](); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewQueue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueue_Clear(t *testing.T) {
	want := &Queue[int]{&sync.RWMutex{}, nil, nil, 0}
	t.Run("Clear empty queue", func(t *testing.T) {
		q := &Queue[int]{&sync.RWMutex{}, nil, nil, 0}
		q.Clear()
		if q.size != 0 {
			t.Errorf("Expected empty queue after Clear, but got size %d", q.size)
		}
		if got := NewQueue[int](); !reflect.DeepEqual(got, want) {
			t.Errorf("NewQueue() = %v, want %v", got, want)
		}
	})

	t.Run("Clear non-empty queue", func(t *testing.T) {
		head := &node[int]{value: 1, next: nil, prev: nil}
		tail := &node[int]{value: 2, next: nil, prev: head}
		head.next = tail
		q := &Queue[int]{&sync.RWMutex{}, head, tail, 2}
		q.Clear()
		if q.size != 0 {
			t.Errorf("Expected empty queue after Clear, but got size %d", q.size)
		}
		if got := NewQueue[int](); !reflect.DeepEqual(got, want) {
			t.Errorf("NewQueue() = %v, want %v", got, want)
		}
	})
}

func TestQueue_Dequeue(t *testing.T) {
	type testCase[T any] struct {
		name string
		q    *Queue[T]
		want T
		ok   bool
		size int
	}
	head := &node[int]{value: 1, next: nil, prev: nil}
	tail := &node[int]{value: 2, next: nil, prev: head}
	head.next = tail
	queueTwoElements := &Queue[int]{&sync.RWMutex{}, head, tail, 2}
	tests := []testCase[int]{
		{"Dequeue empty queue",
			&Queue[int]{&sync.RWMutex{}, nil, nil, 0},
			0, false, 0,
		},
		{"Dequeue one element in queue",
			&Queue[int]{&sync.RWMutex{}, &node[int]{value: 1, next: nil, prev: nil}, nil, 1},
			1, true, 0,
		},
		{"Dequeue two elements in queue",
			queueTwoElements,
			1, true, 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.q.Dequeue()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Dequeue() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.ok {
				t.Errorf("Dequeue() got1 = %v, want %v", got1, tt.ok)
			}
			if tt.q.size != tt.size {
				t.Errorf("Dequeue() size = %v, want %v", tt.q.size, tt.size)
			}
		})
	}

	t.Run("Dequeue three elements in queue", func(t *testing.T) {
		head := &node[int]{value: 1, next: nil, prev: nil}
		head.next = &node[int]{value: 2, next: nil, prev: head}
		tail := &node[int]{value: 3, next: nil, prev: head.next}
		q := &Queue[int]{&sync.RWMutex{}, head, tail, 3}
		got, got1 := q.Dequeue()
		if q.size != 2 {
			t.Errorf("Expected empty queue after Clear, but got size %d", q.size)
		}
		if !reflect.DeepEqual(got, 1) {
			t.Errorf("Dequeue() got = %v, want %v", got, 1)
		}
		if got1 != true {
			t.Errorf("Dequeue() got1 = %v, want %v", got1, true)
		}
	})
}

func TestQueue_Enqueue(t *testing.T) {
	type args[T any] struct {
		value T
	}
	type testCase[T any] struct {
		name string
		q    Queue[T]
		args args[T]
	}
	tests := []testCase[int]{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.q.Enqueue(tt.args.value)
		})
	}
}

func TestQueue_ForEach(t *testing.T) {
	type args[T any] struct {
		callback func(n *node[T])
	}
	type testCase[T any] struct {
		name string
		q    Queue[T]
		args args[T]
	}
	tests := []testCase[int]{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.q.ForEach(tt.args.callback)
		})
	}
}

func TestQueue_IsEmpty(t *testing.T) {
	type testCase[T any] struct {
		name string
		q    Queue[T]
		want bool
	}
	tests := []testCase[int]{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.q.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueue_Peek(t *testing.T) {
	type testCase[T any] struct {
		name  string
		q     Queue[T]
		want  T
		want1 bool
	}
	tests := []testCase[int]{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.q.Peek()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Peek() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Peek() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestQueue_Size(t *testing.T) {
	type testCase[T any] struct {
		name string
		q    Queue[T]
		want int
	}
	tests := []testCase[int]{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.q.Size(); got != tt.want {
				t.Errorf("Size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueue_String(t *testing.T) {
	type testCase[T any] struct {
		name string
		q    Queue[T]
		want string
	}
	tests := []testCase[int]{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.q.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
