// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"
import http "net/http"
import io "io"
import mock "github.com/stretchr/testify/mock"

// HTTPClient is an autogenerated mock type for the HTTPClient type
type HTTPClient struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, uri
func (_m *HTTPClient) Get(ctx context.Context, uri string) (*http.Response, error) {
	ret := _m.Called(ctx, uri)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(context.Context, string) *http.Response); ok {
		r0 = rf(ctx, uri)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, uri)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Post provides a mock function with given fields: ctx, url, contentType, body
func (_m *HTTPClient) Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	ret := _m.Called(ctx, url, contentType, body)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(context.Context, string, string, io.Reader) *http.Response); ok {
		r0 = rf(ctx, url, contentType, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, io.Reader) error); ok {
		r1 = rf(ctx, url, contentType, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
