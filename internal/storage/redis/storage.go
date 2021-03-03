package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type Storage struct {
	client *redis.Client
}

func NewStorage(url string) (*Storage, error) {
	options, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url failed: %s", err)
	}

	client := redis.NewClient(options)

	return &Storage{
		client: client,
	}, nil
}

func (s *Storage) Get(key string) (string, error) {
	value, err := s.client.Get(context.TODO(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return value, nil
		}

		return "", fmt.Errorf("get redis key failed: %s, key: %s", err, key)
	}

	return value, nil
}

func (s *Storage) Set(key, value string) error {
	if _, err := s.client.Set(context.TODO(), key, value, 0).Result(); err != nil {
		if err == redis.Nil {
			return nil
		}

		return fmt.Errorf("set redis key failed: %s, key: %s, value: %s", err, key, value)
	}

	return nil
}
