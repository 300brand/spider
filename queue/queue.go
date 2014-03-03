package queue

import (
	"errors"
)

type Queue interface {
	Dequeue() (string, error)
	Enqueue(string) error
}

var ErrEmpty = errors.New("Queue empty")
