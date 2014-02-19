package storage

import (
	"github.com/300brand/spider/page"
	"net/http"
)

type Storage interface {
	Exists(string) (bool, error)
	Retrieve(string, *page.Page) error
	Store(*page.Page) error
}
