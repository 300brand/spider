package storage

import (
	"launchpad.net/gocheck"
)

type MongoSuite struct {
	Url string
}

var _ = gocheck.Suite(&MongoSuite{
	Url: "localhost/test_backend",
})

func (s *MongoSuite) SetUpTest(c *gocheck.C)    {}
func (s *MongoSuite) TearDownTest(c *gocheck.C) {}

// Runs the standard backend tests
func (s *MongoSuite) TestBackend(c *gocheck.C) {
	b, err := NewMongo(s.Url)
	c.Assert(err, gocheck.IsNil)
	defer b.Close()
	testBackend(c, b)
}
