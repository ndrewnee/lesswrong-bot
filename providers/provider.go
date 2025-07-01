package providers

import (
	"context"

	"github.com/ndrewnee/lesswrong-bot/models"
)

type PostProvider interface {
	GetRandomPost(ctx context.Context) (models.Post, error)
	GetName() string
	GetCacheKey() string
}

type TopPostsProvider interface {
	GetTopPosts(ctx context.Context) (string, error)
	GetName() string
}

// Internal interfaces for providers that may need different signatures
type HTTPClient interface {
	Get(ctx context.Context, url string) (*HTTPResponse, error)
	Post(ctx context.Context, url, contentType string, body interface{}) (*HTTPResponse, error)
}

type HTTPResponse struct {
	StatusCode int
	Body       []byte
}

type Storage interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, expire int) error
}