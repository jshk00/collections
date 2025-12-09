package httpx

import (
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
)

type (
	ResponseHook func(*Response) error
	RequestHook  func(*Request) error
	RetryHook    func(*Request) (*Response, error)
)

type Request struct {
	*http.Request
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
	return &Request{Request: req, queries: make(url.Values), err: err}
}

func (r *Request) SetContext(ctx context.Context) *Request {
	if r.err == nil {
		r.Request = r.WithContext(ctx)
	}
	return r
}

func (r *Request) SetHeader(k, v string) *Request {
	if r.err == nil {
		r.Header.Set(k, v)
	}
	return r
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

func (r *Request) Clone(body io.Reader) *Request {
	r.URL.RawQuery = ""
	req, err := http.NewRequest(r.Method, r.URL.String(), body)
	rr := &Request{
		Request:   req,
		err:       err,
		retryHook: r.retryHook,
	}
	maps.Copy(rr.queries, r.queries)
	maps.Copy(rr.Header, r.Header)
	return rr
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

	if host := r.Header.Get("Host"); host != "" {
		r.Host = host
	}

	// initiate trace once per request if available
	if r.client.trace {
		r.tracer = &TraceInfo{}
		r.Request = r.WithContext(r.tracer.Tracer(r.Context()))
	}

	r.URL.RawQuery = r.queries.Encode()
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
