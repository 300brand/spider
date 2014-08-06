package main

import (
	"flag"
	"github.com/300brand/logger"
	"github.com/PuerkitoBio/fetchbot"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	mu     sync.Mutex
	stores = make(map[string]*MySQL)

	MaxDepth   = flag.Int("maxdepth", 1, "Maximum depth to descend past start page")
	MaxRetries = flag.Int("retries", 3, "Number of retries before succumbing to failure")
	MySQLDSN   = flag.String("mysql", "root:@tcp(localhost:49159)/spider", "MySQL DSN")
)

var (
	_ fetchbot.HandlerFunc = errorHandler
	_ fetchbot.HandlerFunc = getHandler
)

// Intended to run as go dequeue(domain)
func dequeue(bot *fetchbot.Fetcher, filter Filter) {
	queue := bot.Start()
	u, err := url.Parse(filter.Start)
	if err != nil {
		logger.Error.Fatalf("url.Parse: %s", err)
	}
	store, ok := stores[filter.Ident]
	if !ok {
		logger.Error.Fatalf("No store for %s", filter.Ident)
	}

	// Auto re-queue startpoint
	go func() {
		for {
			logger.Info.Printf("[%s] Queuing startpoint %s", filter.Ident, u)
			cmd := &Command{U: u, M: "GET"}
			if err := queue.Send(cmd); err != nil {
				logger.Error.Printf("queue.Send(%#v): %s", cmd, err)
				continue
			}
			logger.Info.Printf("[%s] Startpoint requeue in %s, %s", filter.Ident, filter.Restart, time.Now().Add(filter.Restart))
			<-time.After(filter.Restart)
		}
	}()

	// Dequeue from store; enqueue to bot queue
	for {
		id, rawurl, err := store.Next()
		switch {
		case err == ErrNoNext:
			logger.Trace.Printf("[%s] Nothing in queue; waiting %s", filter.Ident, filter.Restart)
			<-time.After(filter.Restart)
			continue
		case err != nil:
			logger.Error.Printf("[%s] Error fetching next: %s", filter.Ident, err)
			break
		}
		cmd := &Command{
			M:     "GET",
			Id:    id,
			Depth: 1,
		}
		cmd.U, _ = url.Parse(rawurl)
		if err := queue.Send(cmd); err != nil {
			logger.Error.Printf("[%s] Error queuing: %s", filter.Ident, err)
			break
		}
		<-time.After(bot.CrawlDelay)
	}
}

func enqueueLinks(ctx *fetchbot.Context, doc *goquery.Document) {
	mu.Lock()
	defer mu.Unlock()

	cmd, ok := ctx.Cmd.(*Command)
	if !ok {
		logger.Error.Fatalf("ctx.Cmd is not of type Command: %#v", ctx.Cmd)
	}

	filter, ok := filters[cmd.URL().Host]
	if !ok {
		logger.Error.Fatalf("No filter defined for %s", ctx.Cmd.URL().Host)
	}

	store, ok := stores[filter.Ident]
	if !ok {
		logger.Warn.Printf("No store defined for %s, skipping.", filter.Ident)
		return
	}

	doc.Find(filter.CSSSelector).Each(func(i int, s *goquery.Selection) {
		val, _ := s.Attr("href")
		// Resolve address
		u, err := ctx.Cmd.URL().Parse(val)
		if err != nil {
			logger.Error.Printf("[%s] Resolve URL %s - %s", filter.Ident, val, err)
			return
		}

		// Reject
		for _, re := range filter.Reject {
			if re.MatchString(u.Path) {
				logger.Warn.Printf("[%s] REJECT %s", filter.Ident, u)
				return
			}
		}

		// Accept - if none, accept all
		if len(filter.Accept) == 0 {
			logger.Info.Printf("[%s] ACCEPT %s with *", filter.Ident, u)
			if err := store.Enqueue(u.String()); err != nil {
				logger.Error.Printf("[%s] Enqueue head: %s - %s", filter.Ident, u, err)
			}
			return
		}

		// Accept - only accept matching
		for _, re := range filter.Accept {
			if re.MatchString(u.Path) {
				logger.Info.Printf("[%s] ACCEPT %s with %s", filter.Ident, u, re.String())
				if err := store.Enqueue(u.String()); err != nil {
					logger.Error.Printf("[%s] Enqueue head: %s - %s", filter.Ident, u, err)
					return
				}
			}
		}

	})
}

func errorHandler(ctx *fetchbot.Context, res *http.Response, err error) {
	logger.Warn.Printf("%s", err)

	cmd, ok := ctx.Cmd.(*Command)
	if !ok {
		logger.Error.Fatalf("ctx.Cmd is not of type Command: %#v", ctx.Cmd)
	}

	if cmd.Retries >= *MaxRetries {
		logger.Error.Printf("Max retries (%d) for %s %s", *MaxRetries, cmd.Method(), cmd.URL())
		return
	}

	cmd.Retries++
	logger.Debug.Printf("RETRY [%d] %s %s", cmd.Retries, cmd.Method(), cmd.URL())
	if err = ctx.Q.Send(cmd); err != nil {
		logger.Error.Printf("Error requeuing: %s", err)
	}
}

func getHandler(ctx *fetchbot.Context, res *http.Response, err error) {
	switch sc := res.StatusCode; sc {
	case 200:
	default:
		logger.Warn.Printf("ERROR [%d] Leaving for requeue %s", sc, ctx.Cmd.URL())
		return
	}

	cmd, ok := ctx.Cmd.(*Command)
	if !ok {
		logger.Error.Fatalf("ctx.Cmd is not of type Command: %#v", ctx.Cmd)
	}

	// Process the body to find the links
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		logger.Error.Printf("%s %s - %s\n", cmd.Method(), cmd.URL(), err)
		return
	}

	if cmd.Depth < *MaxDepth {
		// Enqueue all links into store
		enqueueLinks(ctx, doc)
		return
	}

	filter, ok := filters[cmd.URL().Host]
	if !ok {
		logger.Error.Fatalf("No filter defined for %s", ctx.Cmd.URL().Host)
	}

	store, ok := stores[filter.Ident]
	if !ok {
		logger.Warn.Printf("No store defined for %s, skipping.", filter.Ident)
		return
	}

	cmd.Title = "No title tag"
	titleNode := doc.Find("title")
	if titleNode.Length() > 0 {
		cmd.Title = titleNode.First().Text()
	}

	if err := store.Save(cmd); err != nil {
		logger.Error.Printf("[%s] Error saving: %s", filter.Ident, err)
	}
}

// logHandler prints the fetch information and dispatches the call to the wrapped Handler.
func logHandler(wrapped fetchbot.Handler) fetchbot.Handler {
	return fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		if err == nil {
			logger.Info.Printf("%s [%d] %s - %s", ctx.Cmd.Method(), res.StatusCode, ctx.Cmd.URL(), res.Header.Get("Content-Type"))
		}
		wrapped.Handle(ctx, res, err)
	})
}

func main() {
	flag.Parse()

	mux := fetchbot.NewMux()

	mux.HandleErrors(fetchbot.HandlerFunc(errorHandler))

	mux.Response().Method("GET").ContentType("text/html").Handler(fetchbot.HandlerFunc(getHandler))

	handler := logHandler(mux)
	// Start queues
	for _, filter := range filters {
		var err error
		if stores[filter.Ident], err = DialMySQL(*MySQLDSN, filter.Ident); err != nil {
			logger.Error.Fatalf("DialMySQL: %s", err)
		}
		logger.Debug.Printf("%+v", stores)
		bot := fetchbot.New(handler)
		go dequeue(bot, filter)
	}
	select {}
}
