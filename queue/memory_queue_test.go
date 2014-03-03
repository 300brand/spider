package queue

import (
	"launchpad.net/gocheck"
)

type MemoryQueueSuite struct{}

var _ = gocheck.Suite(new(MemoryQueueSuite))

func (s *MemoryQueueSuite) TestQueue(c *gocheck.C) {
	q := NewMemoryQueue(3)
	strs := []string{"A", "B", "C"}
	for _, str := range strs {
		c.Assert(q.Enqueue(str), gocheck.IsNil)
	}

	for _, exp := range strs {
		got, err := q.Dequeue()
		c.Assert(err, gocheck.IsNil)
		c.Assert(got, gocheck.Equals, exp)
	}

	got, err := q.Dequeue()
	c.Assert(err, gocheck.Equals, ErrEmpty)
	c.Assert(got, gocheck.Equals, "")
}
