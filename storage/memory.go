package storage

import (
	"github.com/300brand/spider/config"
	"github.com/300brand/spider/page"
)

type Memory struct {
	config config.Config
	pages  map[string]page.Page
}

var _ Storage = new(Memory)

func NewMemory() (m *Memory, err error) {
	m = &Memory{
		pages: make(map[string]page.Page),
	}
	return
}

func (m *Memory) Close() (err error) {
	return
}

func (m *Memory) GetPage(url string, p *page.Page) (err error) {
	if _, ok := m.pages[url]; ok {
		*p = m.pages[url]
		return
	}
	return ErrNotFound
}
func (m *Memory) SavePage(p *page.Page) (err error) {
	m.pages[p.URL] = *p
	return
}
func (m *Memory) GetConfig(c *config.Config) (err error) {
	*c = m.config
	return
}
func (m *Memory) SaveConfig(c *config.Config) (err error) {
	m.config = *c
	return
}
