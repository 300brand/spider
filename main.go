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
	"time"
)

var (
	storeSqlite = flag.String("store.sqlite", "", "Directory to store SQLite files")
	storeMongo  = flag.String("store.mongo", "", "Connection string to mongodb store - host:port/db")
	queueMongo  = flag.String("queue.mongo", "", "Connection string to mongodb queue - host:port/db")
)

func main() {
	flag.Parse()
	var err error

	// Set up storage backend
	var store storage.Storage
	switch {
	case *storeMongo != "":
		if store, err = storage.NewMongo(*storeMongo); err != nil {
			logger.Error.Fatal(err)
		}
	case *storeSqlite != "":
		if store, err = storage.NewSqlite(*storeSqlite); err != nil {
			logger.Error.Fatal(err)
		}
	default:
		store, _ = storage.NewMemory()
	}

	// Set up queue backend
	var q queue.Queue
	switch {
	case *queueMongo != "":
		if q, err = queue.NewMongo(*queueMongo); err != nil {
			logger.Error.Fatal(err)
		}
	default:
		q = queue.NewMemory(128)
	}

	err = store.SaveConfig(&config.Config{
		Domains: []domain.Domain{
			{
				Name:       "300Brand",
				URL:        "http://300brand.com",
				Delay:      time.Second,
				Redownload: time.Hour,
			},
			// {
			// 	Name:       "ASAA",
			// 	URL:        "http://asaa.org",
			// 	Delay:      10 * time.Second,
			// 	Redownload: 12 * time.Hour,
			// },
			// {
			// 	Name:       "Community College Week Magazine",
			// 	URL:        "http://ccweek.com",
			// 	Delay:      10 * time.Second,
			// 	Redownload: 12 * time.Hour,
			// },
			// {
			// 	Name:       "Health IT Security",
			// 	URL:        "http://healthitsecurity.com",
			// 	Delay:      10 * time.Second,
			// 	Redownload: 12 * time.Hour,
			// },
		},
	})
	if err != nil {
		logger.Error.Fatal(err)
	}

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
