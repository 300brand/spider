package scheduler

import (
	"errors"
	"github.com/300brand/logger"
	"github.com/300brand/spider/config"
	"github.com/300brand/spider/domain"
	"github.com/300brand/spider/page"
	"github.com/300brand/spider/queue"
	"github.com/300brand/spider/storage"
	"time"
)

type Scheduler struct {
	config       *config.Config
	curDomain    *domain.Domain
	curUrl       string
	defaultQueue queue.Queue
	err          error
	notify       chan *domain.Domain
	once         bool
	queues       map[string]queue.Queue
	shutdown     chan bool
	store        storage.Storage
}

var (
	ErrQueueNotFound = errors.New("Queue not found")
)

func New(q queue.Queue, store storage.Storage) (s *Scheduler, err error) {
	s = &Scheduler{
		config:       new(config.Config),
		defaultQueue: q,
		store:        store,
	}

	if err = store.GetConfig(s.config); err != nil {
		return
	}

	s.queues = make(map[string]queue.Queue, len(s.config.Domains))
	s.notify = make(chan *domain.Domain, len(s.config.Domains))
	s.shutdown = make(chan bool, len(s.config.Domains)+1)
	s.Start()

	return
}

func (s *Scheduler) Add(url string) (err error) {
	q, ok := s.queues[domain.FromURL(url)]
	if !ok {
		return ErrQueueNotFound
	}
	return q.Enqueue(url)
}

func (s *Scheduler) Cur(d *domain.Domain, p *page.Page) (err error) {
	*d = *s.curDomain

	switch err := s.store.GetPage(s.curUrl, p); err {
	case nil:
		return nil
	case storage.ErrNotFound:
		*p = page.Page{URL: s.curUrl}
		return nil
	default:
		return err
	}

	return
}

func (s *Scheduler) Err() error {
	return s.err
}

func (s *Scheduler) Next() bool {
	if len(s.queues) == 0 {
		return false
	}

	var d *domain.Domain
	for {
		// Wait for the next domain to surface
		select {
		case d = <-s.notify:
		case <-s.shutdown:
			return false
		}

		var url string
		var err error
		// p := new(page.Page)
		for {
			url, err = s.queues[d.Domain()].Dequeue()

			logger.Info.Printf("Got %s from queue", url)
			if err == queue.ErrEmpty {
				if s.once {
					s.Stop()
					return false
				}
				// When the queue is empty, start over from the top
				logger.Warn.Print("restarting")
				s.restart(d)
				continue
			}

			if err != nil {
				s.err = err
				return false
			}

			break

			// if d.IsStartPoint(url) {
			// 	// Break out of fetch-loop and use startpoint as next URL
			// 	logger.Info.Printf("%s is a startpoint", url)
			// 	break
			// }

			// // Look into DB to see if page exists (should probably be a
			// // separate method)
			// switch s.store.GetPage(url, p) {
			// case nil:
			// 	// Page found, look for next page
			// 	logger.Warn.Printf("%s found, refetching from queue", url)
			// 	continue
			// case storage.ErrNotFound:
			// 	// Page not found, use as next URL
			// 	logger.Info.Printf("%s not previously downloaded", url)
			// 	goto ProcessURL
			// default:
			// 	s.err = err
			// 	return false
			// }
		}

		s.curDomain = d
		s.curUrl = url
		return true
	}
	return false
}

func (s *Scheduler) Once() {
	s.once = true
}

func (s *Scheduler) Start() {
	for i := range s.config.Domains {
		d := &s.config.Domains[i]
		s.queues[d.Domain()] = s.defaultQueue.New(d.Domain())
		go s.notifier(d)
	}
}

func (s *Scheduler) Stop() {
	// <= because s.Next() needs to shutdown, too
	for i := 0; i <= len(s.config.Domains); i++ {
		s.shutdown <- true
	}
}

func (s *Scheduler) Update(p *page.Page) (err error) {
	return s.store.SavePage(p)
}

func (s *Scheduler) notifier(d *domain.Domain) {
	s.restart(d)

	for {
		select {
		case <-time.After(d.Delay):
			s.notify <- d
		case <-s.shutdown:
			return
		}
	}
}

func (s *Scheduler) restart(d *domain.Domain) {
	for i := range d.StartPoints {
		s.Add(d.StartPoints[i])
	}
	if len(d.StartPoints) == 0 {
		s.Add(d.GetURL().String())
	}
}
