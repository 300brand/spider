package queue

import (
	"github.com/300brand/logger"
	"github.com/kr/beanstalk"
	"strconv"
	"strings"
	"time"
)

type beanstalkQueue struct {
	enq  beanstalk.Tube
	deq  *beanstalk.TubeSet
	conn *beanstalk.Conn
}

var _ Queue = new(beanstalkQueue)

func NewBeanstalk(url string) (q Queue, err error) {
	bq := new(beanstalkQueue)
	if bq.conn, err = beanstalk.Dial("tcp", url); err != nil {
		return
	}
	q = bq.New("default")
	return
}

func (q *beanstalkQueue) New(name string) Queue {
	name = strings.Replace(name, ".", "_", -1)
	newQueue := &beanstalkQueue{
		enq: beanstalk.Tube{
			Conn: q.conn,
			Name: name,
		},
		deq:  beanstalk.NewTubeSet(q.conn, name),
		conn: q.conn,
	}
	return newQueue
}

func (q *beanstalkQueue) Dequeue() (s string, err error) {
	id, body, err := q.deq.Reserve(250 * time.Millisecond)

	var connErr beanstalk.ConnError
	if ce, ok := err.(beanstalk.ConnError); ok {
		connErr = ce
	}
	switch connErr.Err {
	case nil:
		s = string(body)
		logger.Info.Printf("[%d] <- %s", id, s)
		defer q.conn.Delete(id)
	case beanstalk.ErrTimeout:
		err = ErrEmpty
	default:
		logger.Error.Printf("[%d] <- %t %s", id, err)
	}
	return
}

func (q *beanstalkQueue) Enqueue(s string) (err error) {
	id, err := q.enq.Put([]byte(s), 100, 0, time.Hour*24*7)
	switch err {
	case nil:
		logger.Info.Printf("[%d] -> %s", id, s)
	default:
		logger.Error.Printf("[%d] -> %s", id, err)
	}
	return
}

func (q *beanstalkQueue) Len() int {
	stats, err := q.enq.Stats()
	if err != nil {
		logger.Error.Printf("Stats: %s", err)
	}
	logger.Debug.Printf("%+v", stats)
	n, err := strconv.Atoi(stats["current-jobs-ready"])
	return n
}
