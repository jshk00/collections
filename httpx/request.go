package httpx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type (
	ResponseHook func(*Response) error
	RequestHook  func(*Request) error
	RetryHook    func(*Request) (*Response, error)
)

type Request struct {
	req             *http.Request
	responseHook    ResponseHook
	requestHook     RequestHook
	retryHook       RetryHook
	retryHookCalled bool
	client          *Client
	queries         url.Values
	err             error
	tracer          *TraceInfo
}

func NewRequest(method, uri string, body io.Reader) *Request {
	req, err := http.NewRequest(method, uri, body)
	return &Request{req: req, queries: make(url.Values), err: err}
}

func (r *Request) WithContext(ctx context.Context) *Request {
	if r.err == nil {
		r.req = r.req.WithContext(ctx)
	}
	return r
}

func (r *Request) Context() context.Context {
	return r.req.Context()
}

func (r *Request) Method() string {
	return r.req.Method
}

func (r *Request) GetBodyFn() func() (io.ReadCloser, error) {
	return r.req.GetBody
}

func (r *Request) SetBodyFn(fn func() (io.ReadCloser, error)) *Request {
	r.req.GetBody = fn
	return r
}

func (r *Request) URL() *url.URL {
	return r.req.URL
}

func (r *Request) SetHeader(k, v string) *Request {
	if r.err == nil {
		r.req.Header.Set(k, v)
	}
	return r
}

func (r *Request) SetCookies(c *http.Cookie) *Request {
	r.req.AddCookie(c)
	return r
}

func (r *Request) Raw() *http.Request {
	return r.req
}

func (r *Request) SetHeaders(hdrs map[string]string) *Request {
	if r.err == nil {
		for k, v := range hdrs {
			r.SetHeader(k, v)
		}
	}
	return r
}

func (r *Request) SetQuery(k, v string) *Request {
	r.queries.Set(k, v)
	return r
}

func (r *Request) SetQueries(queries map[string]string) *Request {
	for k, v := range queries {
		r.SetQuery(k, v)
	}
	return r
}

func (r *Request) SetRequestHook(hook RequestHook) *Request {
	r.requestHook = hook
	return r
}

func (r *Request) SetResponseHook(hook ResponseHook) *Request {
	r.responseHook = hook
	return r
}

func (r *Request) SetRetryHook(hook RetryHook) *Request {
	r.retryHook = hook
	return r
}

// Hook execution order:
//
//  1. requestHook — runs before sending the request.
//  2. retryHook   — if defined, takes full control over retries and
//     determines the final response. In this case, responseHook is
//     NOT invoked.
//  3. responseHook — runs only if no retryHook is defined.
//
// Important:
//
//   - If retryHook is defined (custom or default), responseHook will be ignored.
//     This avoids conflicts from reading res.Body multiple times.
//
//   - When using the default retryHook, place any post-processing logic
//     (e.g. decoding JSON, logging, validation) in the PostRecv function itself.
//
//   - When writing a custom retryHook, encapsulate your retry decision and
//     any post-processing logic inside the retryHook implementation.
//
// This ensures hooks remain predictable and prevents accidental multiple
// reads of the response body.
func (r *Request) Exec() (*Response, error) {
	// check if no error in building request
	if r.err != nil {
		return nil, r.err
	}

	if host := r.req.Header.Get("Host"); host != "" {
		r.req.Host = host
	}

	// initiate trace once per request if available
	if r.client.trace {
		r.tracer = &TraceInfo{}
		r.req = r.req.WithContext(r.tracer.Tracer(r.req.Context()))
	}

	r.req.URL.RawQuery = r.queries.Encode()
	if r.requestHook != nil {
		if err := r.requestHook(r); err != nil {
			return nil, fmt.Errorf("failed to execute request hook: %w", err)
		}
	}

	// Excuetion and hooks placement
	if r.retryHook != nil && !r.retryHookCalled {
		r.retryHookCalled = true
		return r.retryHook(r)
	}

	res, err := r.client.exec(r)
	if r.responseHook != nil && r.requestHook == nil {
		if err := r.responseHook(res); err != nil {
			return nil, fmt.Errorf("failed to execute response hook: %w", err)
		}
	}
	return res, err
}
