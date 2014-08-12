package main

import (
	"flag"
	"github.com/300brand/logger"
	"github.com/300brand/spider/rule"
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
	rules  = make(map[string]*rule.Rule)

	MaxDepth   = flag.Int("maxdepth", 1, "Maximum depth to descend past start page")
	MaxRetries = flag.Int("retries", 3, "Number of retries before succumbing to failure")
	MySQLDSN   = flag.String("mysql", "root:@tcp(localhost:49159)/spider", "MySQL DSN")
)

var (
	_ fetchbot.HandlerFunc = errorHandler
	_ fetchbot.HandlerFunc = getHandler
)

// Intended to run as go dequeue(domain)
func dequeue(bot *fetchbot.Fetcher, rule *rule.Rule, kill chan bool) {
	queue := bot.Start()
	u, err := url.Parse(rule.Start)
	if err != nil {
		logger.Error.Fatalf("url.Parse: %s", err)
	}
	store, ok := stores[rule.Ident]
	if !ok {
		logger.Error.Fatalf("No store for %s", rule.Ident)
	}

	for {
		select {
		case <-time.After(rule.Restart):
			// Auto re-queue startpoint
			logger.Info.Printf("[%s] Queuing startpoint %s", rule.Ident, u)
			cmd := &Command{U: u, M: "GET"}
			if err := queue.Send(cmd); err != nil {
				logger.Error.Printf("queue.Send(%#v): %s", cmd, err)
				continue
			}
			logger.Info.Printf("[%s] Startpoint requeue in %s, %s", rule.Ident, rule.Restart, time.Now().Add(rule.Restart))
		case <-time.After(bot.CrawlDelay):
			// Dequeue from store; enqueue to bot queue
			id, rawurl, err := store.Next()
			switch {
			case err == ErrNoNext:
				logger.Trace.Printf("[%s] Nothing in queue; waiting %s", rule.Ident, rule.Restart)
				<-time.After(rule.Restart)
				continue
			case err != nil:
				logger.Error.Printf("[%s] Error fetching next: %s", rule.Ident, err)
				continue
			}
			cmd := &Command{
				M:     "GET",
				Id:    id,
				Depth: 1,
			}
			cmd.U, _ = url.Parse(rawurl)
			if err := queue.Send(cmd); err != nil {
				logger.Error.Printf("[%s] Error queuing: %s", rule.Ident, err)
			}
		case <-kill:
			logger.Trace.Printf("[%s] Shutting down dequeue", rule.Ident)
			queue.Close()
			return
		}
	}
}

func enqueueLinks(ctx *fetchbot.Context, doc *goquery.Document) {
	mu.Lock()
	defer mu.Unlock()

	cmd, ok := ctx.Cmd.(*Command)
	if !ok {
		logger.Error.Fatalf("ctx.Cmd is not of type Command: %#v", ctx.Cmd)
	}

	rule, ok := rules[cmd.URL().Host]
	if !ok {
		logger.Error.Fatalf("No rule defined for %s", cmd.URL().Host)
	}

	store, ok := stores[rule.Ident]
	if !ok {
		logger.Warn.Printf("[%s] No store defined, skipping.", rule.Ident)
		return
	}

	links, err := rule.ExtractLinks(doc, cmd.URL())
	if err != nil {
		logger.Error.Printf("[%s] rule.ExtractLinks(%s): %s", rule.Ident, cmd.URL(), err)
		return
	}

	for _, u := range links.Accept {
		if err := store.Enqueue(u.String()); err != nil {
			logger.Error.Printf("[%s] store.Enqueue(%s): %s", rule.Ident, u, err)
		}
	}
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

	rule, ok := rules[cmd.URL().Host]
	if !ok {
		logger.Error.Fatalf("No rule defined for %s", cmd.URL().Host)
	}

	store, ok := stores[rule.Ident]
	if !ok {
		logger.Warn.Printf("[%s] No store defined, skipping.", rule.Ident)
		return
	}

	cmd.Title = rule.ExtractTitle(doc)

	if err := store.Save(cmd); err != nil {
		logger.Error.Printf("[%s] Error saving: %s", rule.Ident, err)
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

func runQueues(handler fetchbot.Handler) {

	// Start queues
	for _, rule := range rules {
		var err error
		if stores[rule.Ident], err = DialMySQL(*MySQLDSN, rule.Ident); err != nil {
			logger.Error.Fatalf("DialMySQL: %s", err)
		}
		logger.Debug.Printf("%+v", stores)
		bot := fetchbot.New(handler)
		killChan := make(chan bool)
		go dequeue(bot, rule, killChan)
	}
	select {}
}

func main() {
	flag.Parse()

	mux := fetchbot.NewMux()

	mux.HandleErrors(fetchbot.HandlerFunc(errorHandler))

	mux.Response().Method("GET").ContentType("text/html").Handler(fetchbot.HandlerFunc(getHandler))

	handler := logHandler(mux)

	runQueues(handler)
}
