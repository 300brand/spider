package scheduler

import (
	"github.com/300brand/spider/config"
	"github.com/300brand/spider/domain"
	"github.com/300brand/spider/page"
	"github.com/300brand/spider/queue"
	"github.com/300brand/spider/storage"
	"time"
)

type Scheduler struct {
	curDomain    *domain.Domain
	curUrl       string
	defaultQueue queue.Queue
	queues       map[string]queue.Queue
	store        *storage.Storage
	config       *config.Config
	notify       chan *domain.Domain
	shutdown     chan bool
}

func New(q queue.Queue, store *storage.Storage) (s *Scheduler, err error) {
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
	s.shutdown = make(chan bool, len(s.config.Domains))

	return
}

func (s *Scheduler) Add(url string) (err error) {
	tld := domain.FromURL(url)
	return s.queues[tld].Enqueue(url)
}

func (s *Scheduler) Cur(d *domain.Domain, p *page.Page) (err error) {
	d = s.curDomain

	switch err := s.store.GetPage(s.curUrl, p); err {
	case nil, storage.ErrNotFound:
		return nil
	default:
		return err
	}

	return
}

func (s *Scheduler) Next() bool {
	if len(s.queues) == 0 {
		return false
	}
	for {
		select {
		case d := <-s.notify:
			url, err := s.queues[d.Domain()].Dequeue()
			if err != nil {
				// Queue empty..
				continue
			}
			s.curDomain = d
			s.curUrl = url
			return true
		}
	}
	return false
}

func (s *Scheduler) Start() {
	for i := range s.config.Domains {
		d := &s.config.Domains[i]
		s.queues[d.Domain()] = s.defaultQueue.New(d.Domain())
		go s.notifier(d)
	}
}

func (s *Scheduler) Stop() {
	for _ = range s.config.Domains {
		s.shutdown <- true
	}
}

func (s *Scheduler) notifier(d *domain.Domain) {
	for i := range d.StartPoints {
		s.Add(d.StartPoints[i])
	}
	if len(d.StartPoints) == 0 {
		s.Add(d.URL)
	}

	for {
		select {
		case <-time.After(d.Delay):
			s.notify <- d
		case <-s.shutdown:
			return
		}
	}
}
