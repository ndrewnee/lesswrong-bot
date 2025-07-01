package memory

import (
	"context"
	"sync"
	"time"
)

type Storage struct {
	mu    sync.RWMutex
	cache map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		cache: make(map[string]string),
	}
}

func (s *Storage) Get(_ context.Context, key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache[key], nil
}

func (s *Storage) Set(_ context.Context, key, value string, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[key] = value
	return nil
}
