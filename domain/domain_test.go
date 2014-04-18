package domain

import (
	"github.com/300brand/spider/page"
	"github.com/300brand/spider/samplesite"
	"launchpad.net/gocheck"
	"testing"
	"time"
)

type DomainSuite struct{}

var _ = gocheck.Suite(new(DomainSuite))

func Test(t *testing.T) { gocheck.TestingT(t) }

func (s *DomainSuite) TestRobotsTxt(c *gocheck.C) {
	d := &Domain{URL: samplesite.URL}
	tests := map[*page.Page]error{
		&page.Page{URL: samplesite.URL + "/"}:         nil,
		&page.Page{URL: samplesite.URL + "/nospider"}: ErrRobot,
		&page.Page{URL: samplesite.URL + "/article1"}: nil,
		&page.Page{URL: samplesite.URL + "/contact"}:  nil,
	}
	for p, canDL := range tests {
		c.Assert(d.CanDownload(p), gocheck.Equals, canDL)
	}
}

func (s *DomainSuite) TestExcludeRegexp(c *gocheck.C) {
	d := &Domain{
		Exclude: []string{
			"^/cont",
		},
	}
	tests := map[*page.Page]bool{
		&page.Page{URL: samplesite.URL + "/"}:         true,
		&page.Page{URL: samplesite.URL + "/nospider"}: true,
		&page.Page{URL: samplesite.URL + "/article1"}: true,
		&page.Page{URL: samplesite.URL + "/contact"}:  false,
	}
	for p, canDL := range tests {
		c.Assert(d.CanDownload(p) == nil, gocheck.Equals, canDL)
	}
}

func (s *DomainSuite) TestCanLastDownload(c *gocheck.C) {
	d := &Domain{
		URL:        samplesite.URL,
		Redownload: time.Hour,
	}
	tests := map[*page.Page]bool{
		&page.Page{LastDownload: time.Now()}:                        false,
		&page.Page{LastDownload: time.Now().Add(-time.Minute * 59)}: false,
		&page.Page{LastDownload: time.Now().Add(-time.Minute * 61)}: true,
	}

	for p, canDL := range tests {
		c.Assert(d.CanDownload(p) == nil, gocheck.Equals, canDL)
	}
}

func (s *DomainSuite) TestIn_Exclude(c *gocheck.C) {
	d := &Domain{
		URL: samplesite.URL,
		StartPoints: []string{
			samplesite.URL + "/contact",
		},
		Include: []string{
			`^/article\d+`,
		},
		Exclude: []string{
			`^/contact$`,
		},
	}

	tests := map[*page.Page]error{
		&page.Page{URL: samplesite.URL + "/"}:         ErrRegexInclude,
		&page.Page{URL: samplesite.URL + "/nospider"}: ErrRegexInclude,
		&page.Page{URL: samplesite.URL + "/article1"}: nil,
		&page.Page{URL: samplesite.URL + "/contact"}:  nil,
	}
	for p, canDL := range tests {
		c.Assert(d.CanDownload(p), gocheck.Equals, canDL)
	}
}

func (s *DomainSuite) TestDomain(c *gocheck.C) {
	tests := map[*Domain]string{
		&Domain{URL: "http://google.com"}:       "google.com",
		&Domain{URL: "http://www.google.com"}:   "google.com",
		&Domain{URL: "http://blogs.google.com"}: "blogs.google.com",
	}
	for d, exp := range tests {
		c.Assert(d.Domain(), gocheck.Equals, exp)
	}
}
