package providers

import (
	"context"
	"io"
	"time"

	"github.com/ndrewnee/lesswrong-bot/interfaces"
)

type HTTPClientAdapter struct {
	client interfaces.HTTPClient
}

func NewHTTPClientAdapter(client interfaces.HTTPClient) *HTTPClientAdapter {
	return &HTTPClientAdapter{client: client}
}

func (a *HTTPClientAdapter) Get(ctx context.Context, url string) (*HTTPResponse, error) {
	resp, err := a.client.Get(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
	}, nil
}

func (a *HTTPClientAdapter) Post(ctx context.Context, url, contentType string, body interface{}) (*HTTPResponse, error) {
	var reader io.Reader
	if r, ok := body.(io.Reader); ok {
		reader = r
	}

	resp, err := a.client.Post(ctx, url, contentType, reader)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
	}, nil
}

type StorageAdapter struct {
	storage interfaces.Storage
}

func NewStorageAdapter(storage interfaces.Storage) *StorageAdapter {
	return &StorageAdapter{storage: storage}
}

func (a *StorageAdapter) Get(ctx context.Context, key string) (string, error) {
	return a.storage.Get(ctx, key)
}

func (a *StorageAdapter) Set(ctx context.Context, key, value string, expire int) error {
	return a.storage.Set(ctx, key, value, time.Second*time.Duration(expire))
}
