package memory

import (
	"context"
	"time"
)

type Storage struct {
	cache map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		cache: make(map[string]string),
	}
}

func (s *Storage) Get(_ context.Context, key string) (string, error) {
	return s.cache[key], nil
}

func (s *Storage) Set(_ context.Context, key, value string, _ time.Duration) error {
	s.cache[key] = value
	return nil
}
