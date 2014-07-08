package main

import (
	"flag"
	"github.com/300brand/logger"
	"github.com/PuerkitoBio/fetchbot"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"sync"
)

var (
	mu sync.Mutex

	MaxDepth   = flag.Int("maxdepth", 1, "Maximum depth to descend past start page")
	MaxRetries = flag.Int("retries", 3, "Number of retries before succumbing to failure")
)

func main() {
	flag.Parse()

	mux := fetchbot.NewMux()

	mux.HandleErrors(fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
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

	}))

	// Handle GET requests for html responses, to parse the body and enqueue all
	// links as HEAD requests.
	mux.Response().Method("GET").ContentType("text/html").Handler(fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
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
			// Enqueue all links as HEAD requests
			enqueueLinks(ctx, doc)
		}
	}))

	// Handle HEAD requests for html responses coming from the source host - we
	// don't want to crawl links from other hosts.
	mux.Response().Method("HEAD").ContentType("text/html").Handler(fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		cmd, ok := ctx.Cmd.(*Command)
		if !ok {
			logger.Error.Fatalf("ctx.Cmd is not of type Command: %#v", ctx.Cmd)
		}

		cmd.M = "GET"
		if err := ctx.Q.Send(cmd); err != nil {
			logger.Error.Printf("%s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
		}
	}))

	handler := logHandler(mux)
	bot := fetchbot.New(handler)
	queue := bot.Start()
	// Start queue
	for _, cfg := range filters {

		u, err := url.Parse(cfg.Start)
		if err != nil {
			logger.Error.Printf("url.Parse: %s", err)
			return
		}

		cmd := &Command{
			U: u,
			M: "GET",
		}
		if err := queue.Send(cmd); err != nil {
			logger.Error.Printf("queue.Send(%#v): %s", cmd, err)
			return
		}
	}
	queue.Block()
}

// logHandler prints the fetch information and dispatches the call to the wrapped Handler.
func logHandler(wrapped fetchbot.Handler) fetchbot.Handler {
	return fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		if err == nil {
			logger.Info.Printf("%s [%d] %s - %s\n", ctx.Cmd.Method(), res.StatusCode, ctx.Cmd.URL(), res.Header.Get("Content-Type"))
		}
		wrapped.Handle(ctx, res, err)
	})
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

	doc.Find(filter.CSSSelector).Each(func(i int, s *goquery.Selection) {
		val, _ := s.Attr("href")
		// Resolve address
		u, err := ctx.Cmd.URL().Parse(val)
		if err != nil {
			logger.Error.Printf("Resolve URL %s - %s", val, err)
			return
		}

		// logger.Debug.Printf("Found link %s", u.String())

		// Reject
		for _, re := range filter.Reject {
			if re.MatchString(u.Path) {
				logger.Warn.Printf("REJECT %s", u)
				return
			}
		}

		link := &Command{
			U:     u,
			M:     "HEAD",
			Depth: cmd.Depth + 1,
		}

		// Accept - if none, accept all
		if len(filter.Accept) == 0 {
			logger.Info.Printf("ACCEPT %s with *", u)
			if err := ctx.Q.Send(link); err != nil {
				logger.Error.Printf("Enqueue head: %s - %s", u, err)
			}
			return
		}

		// Accept - only accept matching
		for _, re := range filter.Accept {
			if re.MatchString(u.Path) {
				logger.Info.Printf("ACCEPT %s with %s", u, re.String())
				if err := ctx.Q.Send(link); err != nil {
					logger.Error.Printf("Enqueue head: %s - %s", u, err)
					return
				}
			}
		}

	})
}
