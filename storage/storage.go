package storage

import (
	"github.com/300brand/spider/storage/backend"
)

type Storage struct {
	Backend backend.Backend
}

func New(b backend.Backend) (s *Storage) {
	return &Storage{
		Backend: b,
	}
}
