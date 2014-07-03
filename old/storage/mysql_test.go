package storage

import (
	"database/sql"
	"launchpad.net/gocheck"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLSuite struct {
	DSN string
}

var _ = gocheck.Suite(&MySQLSuite{
	DSN: "root:@/test_spider",
})

func (s *MySQLSuite) SetUpTest(c *gocheck.C) {
	db, err := sql.Open("mysql", s.DSN)
	c.Assert(err, gocheck.IsNil)
	db.Exec("DROP TABLE domains, excludes, pages, start_points")
	db.Close()
}

// Runs the standard backend tests
func (s *MySQLSuite) TestBackend(c *gocheck.C) {
	b, err := NewMySQL(s.DSN)
	c.Assert(err, gocheck.IsNil)
	defer b.Close()
	testBackend(c, b)
}
