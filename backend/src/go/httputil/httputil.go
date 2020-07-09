package httputil

import (
	"context"
	"github.com/aws/aws-xray-sdk-go/xray"
	"net/http"
)

type HTTPClientFactory func(ctx context.Context) *http.Client

type contextAppendingTransport struct {
	ctx     context.Context
	wrapped http.RoundTripper
}

func (c *contextAppendingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	withCtx := r.WithContext(c.ctx)
	return c.wrapped.RoundTrip(withCtx)
}

func NewContextAppendingTransport(ctx context.Context, toWrap http.RoundTripper) http.RoundTripper {
	if toWrap == nil {
		toWrap = http.DefaultTransport
	}
	return &contextAppendingTransport{ctx, toWrap}
}

func NewXRAYAwareHTTPClientFactory(baseClient *http.Client) HTTPClientFactory {
	return func(ctx context.Context) *http.Client {
		xrayClient := xray.Client(baseClient)
		xrayClient.Transport = NewContextAppendingTransport(ctx, xrayClient.Transport)
		return xrayClient
	}
}

func DefaultHttpClient(_ context.Context) *http.Client {
	return http.DefaultClient
}