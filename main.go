package main

import (
	"flag"
	"fmt"
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
	mu    sync.Mutex
	rules = make(map[string]*rule.Rule)

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
		logger.Error.Printf("[%s] url.Parse: %s", rule.Ident, err)
		return
	}

	store, err := DialMySQL(*MySQLDSN, rule.Ident)
	if err != nil {
		logger.Error.Printf("[%s] DialMySQL: %s", rule.Ident, err)
		return
	}
	defer store.Close()

	var restart time.Duration
	var crawlDelay = bot.CrawlDelay
	for {
		select {
		case <-time.After(restart):
			restart = rule.Restart
			// Auto re-queue startpoint
			logger.Info.Printf("[%s] Queuing startpoint %s", rule.Ident, u)
			cmd := &Command{U: u, M: "GET"}
			if err := queue.Send(cmd); err != nil {
				logger.Error.Printf("queue.Send(%#v): %s", cmd, err)
				continue
			}
			logger.Info.Printf("[%s] Startpoint requeue in %s, %s", rule.Ident, rule.Restart, time.Now().Add(rule.Restart))
		case <-time.After(crawlDelay):
			crawlDelay = bot.CrawlDelay
			// Dequeue from store; enqueue to bot queue
			id, rawurl, err := store.Next()
			switch {
			case err == ErrNoNext:
				logger.Trace.Printf("[%s] Nothing in queue; waiting %s", rule.Ident, rule.Restart)
				crawlDelay = rule.Restart
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

	store, err := DialMySQL(*MySQLDSN, rule.Ident)
	if err != nil {
		logger.Error.Printf("[%s] DialMySQL: %s", rule.Ident, err)
		return
	}
	defer store.Close()

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

	rule, ok := rules[cmd.URL().Host]
	if !ok {
		logger.Error.Fatalf("No rule defined for %s", cmd.URL().Host)
	}

	if cmd.Depth < rule.MaxDepth {
		// Enqueue all links into store
		enqueueLinks(ctx, doc)
		return
	}

	store, err := DialMySQL(*MySQLDSN, rule.Ident)
	if err != nil {
		logger.Error.Printf("[%s] DialMySQL: %s", rule.Ident, err)
		return
	}
	defer store.Close()

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

func runQueues(handler fetchbot.Handler) (kill chan bool) {
	kill = make(chan bool)
	killers := make(map[string]chan bool, len(rules))
	go func() {
		logger.Debug.Printf("runQueues: Waiting for kill signal")
		<-kill
		logger.Debug.Printf("runQueues: Received kill signal")
		for host, ch := range killers {
			logger.Debug.Printf("runQueues: Killing %s", host)
			ch <- true
		}
	}()
	// Start queues
	logger.Debug.Printf("runQueue: len(rules) = %d", len(rules))
	for host, rule := range rules {
		logger.Debug.Printf("runQueues: Starting %s", host)
		killers[host] = make(chan bool)
		bot := fetchbot.New(handler)
		bot.HttpClient = new(HTTPClient)
		logger.Debug.Printf("runQueues: go dequeue %s", host)
		go dequeue(bot, rule, killers[host])
	}
	return
}

func setupRules(dsn string) (changed bool, newRules map[string]*rule.Rule, err error) {
	defer func() {
		if changed = newRules != nil; !changed {
			newRules = rules
		}
	}()
	c, err := DialConfig(*MySQLDSN)
	if err != nil {
		// At least use the existing rules
		err = fmt.Errorf("DialConfig: %s", err)
		return
	}
	defer c.Close()
	newRules, err = c.Rules()
	return
}

func main() {
	flag.Parse()

	mux := fetchbot.NewMux()

	mux.HandleErrors(fetchbot.HandlerFunc(errorHandler))

	mux.Response().Method("GET").ContentType("text/html").Handler(fetchbot.HandlerFunc(getHandler))

	handler := logHandler(mux)

	var killChan chan bool
	for {
		changed, newRules, err := setupRules(*MySQLDSN)
		if err != nil || newRules == nil {
			logger.Error.Printf("setupRules: %s", err)
			retry := 10 * time.Second
			logger.Error.Printf("Waiting %s to retry", retry)
			<-time.After(retry)
			continue
		}

		if changed {
			rules = newRules
			logger.Info.Printf("Rule changes detected")
			logger.Debug.Printf("New Rules: %v", rules)
			if killChan != nil {
				logger.Info.Printf("Killing spider")
				killChan <- true
				<-time.After(5 * time.Second)
			}
			killChan = runQueues(handler)
			logger.Info.Printf("Queues started")
		}
		<-time.After(30 * time.Second)
	}
}
