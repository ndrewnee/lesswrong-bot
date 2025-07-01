package interfaces

import (
	"context"
	"io"
	"net/http"
	"time"
)

// HTTPClient defines the interface for making HTTP requests
type HTTPClient interface {
	Get(ctx context.Context, url string) (*http.Response, error)
	Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error)
}

// Storage defines the interface for caching operations
type Storage interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, expire time.Duration) error
}