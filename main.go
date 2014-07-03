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
)

func main() {
	flag.Parse()

	u, err := url.Parse(flag.Arg(0))
	if err != nil {
		logger.Error.Printf("url.Parse: %s", err)
		return
	}

	mux := fetchbot.NewMux()

	mux.HandleErrors(fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		logger.Error.Printf("%s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
	}))

	// Handle GET requests for html responses, to parse the body and enqueue all
	// links as HEAD requests.
	mux.Response().Method("GET").ContentType("text/html").Handler(fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		// Process the body to find the links
		doc, err := goquery.NewDocumentFromResponse(res)
		if err != nil {
			logger.Error.Printf("%s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
			return
		}
		// Enqueue all links as HEAD requests
		enqueueLinks(ctx, doc)
	}))

	// Handle HEAD requests for html responses coming from the source host - we
	// don't want to crawl links from other hosts.
	mux.Response().Method("HEAD").Host(u.Host).ContentType("text/html").Handler(fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		if _, err := ctx.Q.SendStringGet(ctx.Cmd.URL().String()); err != nil {
			logger.Error.Printf("%s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
		}
	}))

	handler := logHandler(mux)
	bot := fetchbot.New(handler)
	queue := bot.Start()
	// Start queue
	if _, err := queue.SendStringGet(u.String()); err != nil {
		logger.Error.Printf("queue.SendStringGet(%s): %s", u, err)
		return
	}
	queue.Block()
}

// logHandler prints the fetch information and dispatches the call to the wrapped Handler.
func logHandler(wrapped fetchbot.Handler) fetchbot.Handler {
	return fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		if err == nil {
			logger.Info.Printf("[%d] %s %s - %s\n", res.StatusCode, ctx.Cmd.Method(), ctx.Cmd.URL(), res.Header.Get("Content-Type"))
		}
		wrapped.Handle(ctx, res, err)
	})
}

func enqueueLinks(ctx *fetchbot.Context, doc *goquery.Document) {
	mu.Lock()
	defer mu.Unlock()
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		val, _ := s.Attr("href")
		// Resolve address
		u, err := ctx.Cmd.URL().Parse(val)
		if err != nil {
			logger.Error.Printf("Resolve URL %s - %s", val, err)
			return
		}
		logger.Debug.Printf("Found link %s", u.String())

		// if !dup[u.String()] {
		// if _, err := ctx.Q.SendStringHead(u.String()); err != nil {
		// 	logger.Error.Printf("Enqueue head: %s - %s", u, err)
		// 	// 	} else {
		// 	// 		dup[u.String()] = true
		// }
		// }
	})
}
