package queue

import (
	"sync"
)

type memQueue struct {
	Queue []string
	mutex sync.Mutex
}

var _ Queue = new(memQueue)

func NewMemoryQueue(prealloc int) (q *memQueue) {
	return &memQueue{
		Queue: make([]string, 0, prealloc),
	}
}

func (q *memQueue) Dequeue() (s string, err error) {
	if len(q.Queue) == 0 {
		return "", ErrEmpty
	}
	q.mutex.Lock()
	defer q.mutex.Unlock()
	s, q.Queue = q.Queue[0], q.Queue[1:]
	return
}

func (q *memQueue) Enqueue(s string) (err error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.Queue = append(q.Queue, s)
	return
}
