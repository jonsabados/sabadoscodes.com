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
	return &contextAppendingTransport{ctx, toWrap}
}

func NewXRAYAwareHTTPClientFactory(baseTransport http.RoundTripper) HTTPClientFactory {
	return func(ctx context.Context) *http.Client {
		transport := NewContextAppendingTransport(ctx, baseTransport)
		return xray.Client(&http.Client{
			Transport: transport,
		})
	}
}

func DefaultHttpClient(_ context.Context) *http.Client {
	return http.DefaultClient
}