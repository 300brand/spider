package storage

import (
	"github.com/300brand/spider/config"
	"github.com/300brand/spider/domain"
	"github.com/300brand/spider/page"
	"launchpad.net/gocheck"
	"testing"
	"time"
)

func Test(t *testing.T) { gocheck.TestingT(t) }

func testBackend(c *gocheck.C, s Storage) {
	// Test config in/out
	cfg := new(config.Config)

	c.Assert(s.GetConfig(cfg), gocheck.IsNil)

	cfg.Domains = append(cfg.Domains, domain.Domain{
		Name: "Google",
		URL:  "http://google.com",
		Exclude: []string{
			"/exclude",
			"/(do not|dont)_index_me",
		},
		StartPoints: []string{
			"http://google.com/",
			"http://google.com/starthere",
		},
		Delay: time.Minute,
	})
	c.Assert(s.SaveConfig(cfg), gocheck.IsNil)

	outCfg := new(config.Config)
	c.Assert(s.GetConfig(outCfg), gocheck.IsNil)
	c.Assert(len(outCfg.Domains), gocheck.Equals, len(cfg.Domains))
	c.Assert(len(outCfg.Domains[0].Exclude), gocheck.Equals, len(cfg.Domains[0].Exclude))
	c.Assert(len(outCfg.Domains[0].StartPoints), gocheck.Equals, len(cfg.Domains[0].StartPoints))
	c.Assert(outCfg.Domains[0].Name, gocheck.Equals, cfg.Domains[0].Name)
	c.Assert(outCfg.Domains[0].URL, gocheck.Equals, cfg.Domains[0].URL)

	// Test page in/out
	url := "http://google.com/news.html"

	p := new(page.Page)
	c.Assert(s.GetPage(url, p), gocheck.Equals, ErrNotFound)
	c.Assert(p.URL, gocheck.Equals, "")

	p.URL = url
	c.Assert(s.SavePage(p), gocheck.IsNil)

	*p = page.Page{}
	c.Assert(s.GetPage(url, p), gocheck.IsNil)
	c.Assert(p.URL, gocheck.Equals, url)
}
