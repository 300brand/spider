package queue

import (
	"launchpad.net/gocheck"
)

type MemoryQueueSuite struct{}

var _ = gocheck.Suite(new(MemoryQueueSuite))

func (s *MemoryQueueSuite) TestQueue(c *gocheck.C) {
	q := NewMemoryQueue(3)
	testQueue(c, q)
}
