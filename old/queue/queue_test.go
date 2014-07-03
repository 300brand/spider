package queue

import (
	"launchpad.net/gocheck"
	"testing"
)

func Test(t *testing.T) { gocheck.TestingT(t) }

func testQueue(c *gocheck.C, q Queue) {
	strs := []string{"A", "B", "C"}

	// Fill queue
	for i, str := range strs {
		c.Assert(q.Enqueue(str), gocheck.IsNil)
		c.Assert(q.Len(), gocheck.Equals, i+1)
	}

	// Make a new subqueue
	sub := q.New("subqueue")
	got, err := sub.Dequeue()
	c.Assert(err, gocheck.Equals, ErrEmpty)
	c.Assert(got, gocheck.Equals, "")

	// Empty queue
	for i, exp := range strs {
		got, err = q.Dequeue()
		c.Assert(err, gocheck.IsNil)
		c.Assert(got, gocheck.Equals, exp)
		c.Assert(q.Len(), gocheck.Equals, len(strs)-(i+1))
	}

	got, err = q.Dequeue()
	c.Assert(err, gocheck.Equals, ErrEmpty)
	c.Assert(got, gocheck.Equals, "")
}
