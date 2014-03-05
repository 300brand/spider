package scheduler

import (
	"github.com/300brand/spider/config"
	"github.com/300brand/spider/domain"
	"github.com/300brand/spider/page"
	"github.com/300brand/spider/queue"
	"github.com/300brand/spider/samplesite"
	"github.com/300brand/spider/storage"
	"github.com/300brand/spider/storage/backend"
	"launchpad.net/gocheck"
	"testing"
	"time"
)

type SchedulerSuite struct{}

var _ = gocheck.Suite(new(SchedulerSuite))

func Test(t *testing.T) { gocheck.TestingT(t) }

func (s *SchedulerSuite) TestCrawl(c *gocheck.C) {
	storeBackend, err := backend.NewMemory()
	c.Assert(err, gocheck.IsNil)
	store := storage.New(storeBackend)
	defer store.Close()

	store.SaveConfig(&config.Config{
		Domains: []domain.Domain{
			{
				Name: "Samplesite",
				URL:  samplesite.URL,
			},
		},
	})

	sch, err := New(queue.NewMemoryQueue(1024), store)
	c.Assert(err, gocheck.IsNil)
	sch.Once()
	c.Logf("Scheduler ready, domains: %d", len(sch.config.Domains))

	var p page.Page
	var d domain.Domain
	for sch.Next() {
		c.Check(sch.Cur(&d, &p), gocheck.IsNil)
		c.Logf("Domain: %s Page: %s LastDownload: %s", d.URL, p.URL, p.LastDownload)

		if !d.CanDownload(&p) {
			c.Logf("\tShould not download, skipping!")
			continue
		}

		if p.LastDownload.After(time.Now().Add(-time.Hour)) {
			c.Logf("\tToo soon to re-download page, skipping!")
			continue
		}

		switch err := p.Download(); err {
		case nil:
		case page.ErrNotModified:
			continue
		default:
			c.Error(err)
		}

		c.Assert(sch.Update(&p), gocheck.IsNil)

		links, err := p.Links()
		c.Check(err, gocheck.IsNil)
		c.Logf("\tChecksum: %d Links: %+v", p.Checksum, links)
		for i := range links {
			sch.Add(links[i])
		}
	}
}
