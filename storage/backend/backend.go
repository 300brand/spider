package backend

import (
	"errors"
	"github.com/300brand/spider/config"
	"github.com/300brand/spider/page"
)

var NotFound = errors.New("Not found")

type Backend interface {
	Close() error
	GetPage(url string, p *page.Page) error
	SavePage(p *page.Page) error
	GetConfig(c *config.Config) error
	SaveConfig(c *config.Config) error
}
