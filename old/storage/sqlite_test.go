package storage

import (
	"launchpad.net/gocheck"
	"os"
	"path/filepath"
)

type SqliteSuite struct {
	Dir string
}

var _ = gocheck.Suite(&SqliteSuite{
	Dir: filepath.Join(os.TempDir(), "testdb"),
})

func (s *SqliteSuite) SetUpTest(c *gocheck.C)    { os.RemoveAll(s.Dir) }
func (s *SqliteSuite) TearDownTest(c *gocheck.C) { os.RemoveAll(s.Dir) }

// Runs the standard backend tests
func (s *SqliteSuite) TestBackend(c *gocheck.C) {
	b, err := NewSqlite(s.Dir)
	c.Assert(err, gocheck.IsNil)
	defer b.Close()
	testBackend(c, b)
}
