package page

import (
	"github.com/300brand/spider/samplesite"
	"launchpad.net/gocheck"
	"testing"
)

type PageSuite struct{}

var _ = gocheck.Suite(new(PageSuite))

func Test(t *testing.T) { gocheck.TestingT(t) }

func (s *PageSuite) TestRobotsTxt(c *gocheck.C) {
	p := New(samplesite.URL)
	c.Assert(p.Download(), gocheck.IsNil)
	c.Assert(p.SetTitle(), gocheck.IsNil)
	c.Assert(p.Title, gocheck.Equals, "Index")
}
