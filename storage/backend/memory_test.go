package backend

import (
	"launchpad.net/gocheck"
)

type MemorySuite struct{}

var _ = gocheck.Suite(new(MemorySuite))

// Runs the standard backend tests
func (s *MemorySuite) TestBackend(c *gocheck.C) {
	b, err := NewMemory()
	c.Assert(err, gocheck.IsNil)
	defer b.Close()
	testBackend(c, b)
}
