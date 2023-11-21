package unibus

import (
	"errors"
)

// DebugQueue represents a simple FIFO queue
type DebugQueue struct {
	items   []string
	size    int // Current number of elements in the queue
	maxSize int
}

// NewQueue creates a new empty queue.
func NewQueue(maxSize int) *DebugQueue {
	q := &DebugQueue{}
	q.maxSize = maxSize
	return q
}

// Enqueue adds an item to the rear of the queue.
func (q *DebugQueue) Enqueue(item string) {

	if q.size == q.maxSize {
		q.Dequeue()
	}

	q.items = append(q.items, item)
	q.size++
}

// Dequeue removes and returns the item from the front of the queue.
func (q *DebugQueue) Dequeue() (string, error) {
	if q.size == 0 {
		return "", errors.New("queue is empty")
	}
	frontItem := q.items[0]
	q.items = q.items[1:]
	q.size--
	return frontItem, nil
}

// IsEmpty checks if the queue is empty.
func (q *DebugQueue) IsEmpty() bool {
	return q.size == 0
}
