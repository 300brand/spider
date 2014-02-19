package queue

type Queue interface {
	Dequeue() (string, error)
	Enqueue(string) error
}
