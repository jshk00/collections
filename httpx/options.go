package httpx

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type (
	ResponseHook func(*http.Request, *http.Response) error
	RequestHook  func(*http.Request) error
	RetryHook    func(*http.Request, *http.Response, *http.Client, error) (*http.Response, error)
)

type HTTPOptions struct {
	headers      http.Header
	queries      url.Values
	responseHook ResponseHook
	requestHook  RequestHook
	retryHook    RetryHook
	client       *Client
	ctx          context.Context
	url          string
	body         io.Reader
	method       string
	isTrace      bool
}

func NewHTTPOptions() *HTTPOptions {
	return &HTTPOptions{
		headers: make(http.Header),
		queries: make(url.Values),
	}
}

func (ho *HTTPOptions) Method(method string) *HTTPOptions {
	ho.method = method
	return ho
}

func (ho *HTTPOptions) URL(uri string) *HTTPOptions {
	ho.url = uri
	return ho
}

func (ho *HTTPOptions) Context(ctx context.Context) *HTTPOptions {
	ho.ctx = ctx
	return ho
}

func (ho *HTTPOptions) EnableTrace() *HTTPOptions {
	ho.isTrace = true
	return ho
}

func (ho *HTTPOptions) Body(body io.Reader) *HTTPOptions {
	ho.body = body
	return ho
}

func (ho *HTTPOptions) Header(k, v string) *HTTPOptions {
	ho.headers.Set(k, v)
	return ho
}

func (ho *HTTPOptions) Headers(hdrs map[string]string) *HTTPOptions {
	for k, v := range hdrs {
		ho.Header(k, v)
	}
	return ho
}

func (ho *HTTPOptions) Query(k, v string) *HTTPOptions {
	ho.queries.Set(k, v)
	return ho
}

func (ho *HTTPOptions) Queries(queries map[string]string) *HTTPOptions {
	for k, v := range queries {
		ho.Query(k, v)
	}
	return ho
}

func (ho *HTTPOptions) RequestHook(hook RequestHook) *HTTPOptions {
	ho.requestHook = hook
	return ho
}

func (ho *HTTPOptions) ResponseHook(hook ResponseHook) *HTTPOptions {
	ho.responseHook = hook
	return ho
}

func (ho *HTTPOptions) RetryHook(hook RetryHook) *HTTPOptions {
	ho.retryHook = hook
	return ho
}

func (ho *HTTPOptions) Exec() (*http.Response, error) {
	return ho.client.Exec(ho)
}
