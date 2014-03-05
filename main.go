package main

import (
	"flag"
	"github.com/300brand/logger"
	"github.com/300brand/spider/config"
	"github.com/300brand/spider/domain"
	"github.com/300brand/spider/page"
	"github.com/300brand/spider/queue"
	"github.com/300brand/spider/scheduler"
	"github.com/300brand/spider/storage"
	"github.com/300brand/spider/storage/backend"
	"time"
)

var (
	sqliteDir = flag.String("sqlite.dir", "spider_data/sqlite", "Directory to store SQLite files")
)

func main() {
	flag.Parse()

	bkend, err := backend.NewSqlite(*sqliteDir)
	if err != nil {
		logger.Error.Fatal(err)
	}
	defer bkend.Close()

	store := storage.New(bkend)

	err = store.SaveConfig(&config.Config{
		Domains: []domain.Domain{
			// {
			// 	Name:       "300Brand",
			// 	URL:        "http://300brand.com",
			// 	Delay:      time.Second,
			// 	Redownload: time.Hour,
			// },
			{
				Name:       "ASAA",
				URL:        "http://asaa.org",
				Delay:      10 * time.Second,
				Redownload: 3 * time.Hour,
			},
		},
	})
	if err != nil {
		logger.Error.Fatal(err)
	}

	q := queue.NewMemoryQueue(128)

	sch, err := scheduler.New(q, store)
	if err != nil {
		logger.Error.Fatal(err)
	}

	p, d := new(page.Page), new(domain.Domain)
	for sch.Next() {
		if err := sch.Cur(d, p); err != nil {
			logger.Error.Fatal(err)
		}
		logger.Debug.Printf("Processing: %s [%s]", p.URL, p.LastDownload)

		if !d.CanDownload(p) {
			logger.Warn.Printf("Cannot download %s", p.URL)
			continue
		}

		switch err := p.Download(); err {
		case nil:
			sch.Update(p)
		case page.ErrNotModified:
			logger.Warn.Printf("Not modified: %s", p.URL)
			sch.Update(p)
			continue
		default:
			logger.Error.Printf("Error downloading: %s", err)
			continue
		}

		links, err := p.Links()
		if err != nil {
			logger.Error.Fatal(err)
		}
		for i := range links {
			if err := sch.Add(links[i]); err != nil {
				logger.Warn.Printf("Error adding %s: %s", links[i], err)
				continue
			}
			logger.Trace.Printf("New Link: %s", links[i])
		}
	}

	if err := sch.Err(); err != nil {
		logger.Error.Fatal(err)
	}
}
