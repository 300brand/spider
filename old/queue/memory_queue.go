package queue

import (
	"sync"
)

type memQueue struct {
	Name  string
	Queue []string
	mutex sync.Mutex
}

var _ Queue = new(memQueue)

func NewMemory(prealloc int) (q *memQueue) {
	return &memQueue{
		Name:  "default",
		Queue: make([]string, 0, prealloc),
	}
}

func (q *memQueue) New(name string) Queue {
	newQueue := &memQueue{
		Name:  name,
		Queue: make([]string, 0, len(q.Queue)),
	}
	return newQueue
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
	for i := range q.Queue {
		if q.Queue[i] == s {
			return ErrExists
		}
	}
	q.Queue = append(q.Queue, s)
	return
}

func (q *memQueue) Len() int {
	return len(q.Queue)
}
