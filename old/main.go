package main

import (
	"encoding/json"
	"flag"
	"github.com/300brand/logger"
	"github.com/300brand/spider/config"
	"github.com/300brand/spider/domain"
	"github.com/300brand/spider/feed"
	"github.com/300brand/spider/page"
	"github.com/300brand/spider/queue"
	"github.com/300brand/spider/scheduler"
	"github.com/300brand/spider/storage"
	"log"
	"net/http"
	"os"
)

var (
	storeSqlite     = flag.String("store.sqlite", "", "Directory to store SQLite files")
	storeMongo      = flag.String("store.mongo", "", "Connection string to mongodb store - host:port/db")
	storeMongoShard = flag.Bool("store.mongo.shard", false, "Shard new mongo collections")
	storeMysql      = flag.String("store.mysql", "", "Connection string to mongodb store - user:pass@host:port/db")
	queueBeanstalk  = flag.String("queue.beanstalk", "", "Connection string to beanstalkd queue - host:port")
	queueMongo      = flag.String("queue.mongo", "", "Connection string to mongodb queue - host:port/db")
	queueMongoShard = flag.Bool("queue.mongo.shard", false, "Shard new mongo collections")
	once            = flag.Bool("once", false, "Only crawl sites once, then stop")
	listen          = flag.String("listen", ":8084", "Address:port to listen for HTTP requests")
	printConf       = flag.Bool("printconfig", false, "Print configuration from store and exit")
	rssOnly         = flag.Bool("rssonly", false, "Only run the web interface for RSS exports (don't spider)")
)

func init() {
	logger.Debug = log.New(os.Stdout, "  DEBUG ", logger.DefaultFlags)
	logger.Error = log.New(os.Stderr, "  ERROR ", logger.DefaultFlags)
	logger.Info = log.New(os.Stdout, "   INFO ", logger.DefaultFlags)
	logger.Trace = log.New(os.Stdout, "  TRACE ", logger.DefaultFlags)
	logger.Warn = log.New(os.Stdout, "   WARN ", logger.DefaultFlags)
}

func main() {
	flag.Parse()
	var err error

	// Set up storage backend
	var store storage.Storage
	switch {
	case *storeMysql != "":
		if store, err = storage.NewMySQL(*storeMysql); err != nil {
			logger.Error.Fatal(err)
		}
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

	if *printConf {
		c := new(config.Config)
		if err := store.GetConfig(c); err != nil {
			logger.Error.Fatalf("Error getting config: %s", err)
		}

		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(c); err != nil {
			logger.Error.Fatalf("Error encoding config: %s", err)
		}
		return
	}

	// Set up queue backend
	var q queue.Queue
	switch {
	case *queueBeanstalk != "":
		if q, err = queue.NewBeanstalk(*queueBeanstalk); err != nil {
			logger.Error.Fatal(err)
		}
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

	if *rssOnly {
		select {}
	}

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

		if err := d.CanDownload(p); err != nil {
			logger.Warn.Printf("Cannot download %s: %s", p.URL, err)
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
			l := page.New(links[i])

			if err := d.CanDownload(l); err != nil {
				continue
			}

			if store.GetPage(links[i], new(page.Page)) != storage.ErrNotFound {
				logger.Warn.Printf("Already downloaded %s", links[i])
				continue
			}

			if err := sch.Add(links[i]); err != nil {
				logger.Warn.Printf("Error adding %s: %s", links[i], err)
				continue
			}

			if err := sch.Update(l); err != nil {
				logger.Warn.Printf("Error updating %s: %s", links[i], err)
				continue
			}

			logger.Trace.Printf("New Link: %s", links[i])
		}
	}

	if err := sch.Err(); err != nil {
		logger.Error.Fatal(err)
	}
}
