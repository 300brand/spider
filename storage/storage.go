package storage

import (
	"errors"
	"github.com/300brand/spider/config"
	"github.com/300brand/spider/page"
	"github.com/300brand/spider/storage/backend"
)

type Storage struct {
	Backend backend.Backend
}

var ErrNotFound = errors.New("Not found")

func New(b backend.Backend) (s *Storage) {
	return &Storage{
		Backend: b,
	}
}

func (s *Storage) Close() (err error)                           { return s.Backend.Close() }
func (s *Storage) GetPage(url string, p *page.Page) (err error) { return s.Backend.GetPage(url, p) }
func (s *Storage) SavePage(p *page.Page) (err error)            { return s.Backend.SavePage(p) }
func (s *Storage) GetConfig(c *config.Config) (err error)       { return s.Backend.GetConfig(c) }
func (s *Storage) SaveConfig(c *config.Config) (err error)      { return s.Backend.SaveConfig(c) }
