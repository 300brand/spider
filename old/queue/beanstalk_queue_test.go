package queue

import (
	"launchpad.net/gocheck"
)

type BeanstalkQueueSuite struct {
	Url string
}

var _ = gocheck.Suite(&BeanstalkQueueSuite{
	Url: "localhost:11301",
})

func (s *BeanstalkQueueSuite) SetUpTest(c *gocheck.C) {

}

func (s *BeanstalkQueueSuite) TestQueue(c *gocheck.C) {
	q, err := NewBeanstalk(s.Url)
	c.Assert(err, gocheck.IsNil)
	testQueue(c, q)
}
