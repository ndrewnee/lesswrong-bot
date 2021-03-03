package bot

import (
	"context"
	"io"
	"net/http"
)

type DefaultHTTPClient struct {
	*http.Client
}

func NewHTTPClient() *DefaultHTTPClient {
	return &DefaultHTTPClient{
		Client: http.DefaultClient,
	}
}

func (c *DefaultHTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *DefaultHTTPClient) Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	return c.Do(req)
}
