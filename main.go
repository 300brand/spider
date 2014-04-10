package main

import (
	"flag"
	"github.com/300brand/logger"
	"github.com/300brand/spider/domain"
	"github.com/300brand/spider/feed"
	"github.com/300brand/spider/page"
	"github.com/300brand/spider/queue"
	"github.com/300brand/spider/scheduler"
	"github.com/300brand/spider/storage"
	"log"
	"net/http"
)

var (
	storeSqlite     = flag.String("store.sqlite", "", "Directory to store SQLite files")
	storeMongo      = flag.String("store.mongo", "", "Connection string to mongodb store - host:port/db")
	storeMongoShard = flag.Bool("store.mongo.shard", false, "Shard new mongo collections")
	queueMongo      = flag.String("queue.mongo", "", "Connection string to mongodb queue - host:port/db")
	queueMongoShard = flag.Bool("queue.mongo.shard", false, "Shard new mongo collections")
	once            = flag.Bool("once", false, "Only crawl sites once, then stop")
	listen          = flag.String("listen", ":8084", "Address:port to listen for HTTP requests")
)

func init() {
	logger.Error = log.New(logger.NewColorStderr("r"), "  ERROR ", logger.DefaultFlags)
}

func main() {
	flag.Parse()
	var err error

	// Set up storage backend
	var store storage.Storage
	switch {
	case *storeMongo != "":
		if store, err = storage.NewMongo(*storeMongo, *storeMongoShard); err != nil {
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
		if q, err = queue.NewMongo(*queueMongo, *queueMongoShard); err != nil {
			logger.Error.Fatal(err)
		}
	default:
		q = queue.NewMemory(128)
	}

	http.Handle("/rss/", feed.New(store))
	go func() {
		if err := http.ListenAndServe(*listen, nil); err != nil {
			logger.Error.Fatal(err)
		}
	}()

	sch, err := scheduler.New(q, store)
	if err != nil {
		logger.Error.Fatal(err)
	}
	if *once {
		sch.Once()
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
			if err := p.SetTitle(); err != nil {
				logger.Warn.Printf("Error setting title: %s", err)
			}
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
