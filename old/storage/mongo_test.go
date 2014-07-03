package storage

import (
	"labix.org/v2/mgo"
	"launchpad.net/gocheck"
)

type MongoSuite struct {
	Url string
}

var _ = gocheck.Suite(&MongoSuite{
	Url: "localhost/test_backend",
})

func (s *MongoSuite) SetUpTest(c *gocheck.C) {
	sess, err := mgo.Dial(s.Url)
	c.Assert(err, gocheck.IsNil)
	for _, coll := range []string{"config", "pages_google_com"} {
		_, err = sess.DB("").C(coll).RemoveAll(nil)
		c.Assert(err, gocheck.IsNil)
	}
}

// Runs the standard backend tests
func (s *MongoSuite) TestBackend(c *gocheck.C) {
	b, err := NewMongo(s.Url, false)
	c.Assert(err, gocheck.IsNil)
	defer b.Close()
	testBackend(c, b)
}
