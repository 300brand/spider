package queue

import (
	"errors"
)

type Queue interface {
	New(name string) Queue
	Dequeue() (v string, err error)
	Enqueue(v string) (err error)
	Len() int
}

var ErrEmpty = errors.New("Queue empty")
