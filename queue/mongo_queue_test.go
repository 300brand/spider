package queue

import (
	"labix.org/v2/mgo"
	"launchpad.net/gocheck"
)

type MongoQueueSuite struct {
	Url string
}

var _ = gocheck.Suite(&MongoQueueSuite{
	Url: "localhost/test_queue",
})

func (s *MongoQueueSuite) SetUpTest(c *gocheck.C) {
	sess, err := mgo.Dial(s.Url)
	c.Assert(err, gocheck.IsNil)
	for _, coll := range []string{"default", "subqueue"} {
		_, err = sess.DB("").C("queue_" + coll).RemoveAll(nil)
		c.Assert(err, gocheck.IsNil)
	}
}

func (s *MongoQueueSuite) TestQueue(c *gocheck.C) {
	q, err := NewMongo("localhost/test_queue")
	c.Assert(err, gocheck.IsNil)
	testQueue(c, q)
}
