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
	RetryHook    func(*Request, *Client) (*Response, error)
)

type Request struct {
	*http.Request
	responseHook ResponseHook
	requestHook  RequestHook
	retryHook    RetryHook
	client       *Client
	qrs          url.Values
	err          error
	ok           bool
	tracer       *TraceInfo
}

func NewRequest(method, uri string, body io.Reader) *Request {
	req, err := http.NewRequest(method, uri, body)
	return &Request{Request: req, qrs: make(url.Values), err: err, ok: err == nil}
}

func (r *Request) SetContext(ctx context.Context) *Request {
	if r.ok {
		r.Request = r.WithContext(ctx)
	}
	return r
}

func (r *Request) SetHeader(k, v string) *Request {
	if r.ok {
		r.Header.Set(k, v)
	}
	return r
}

func (r *Request) SetHeaders(hdrs map[string]string) *Request {
	if r.ok {
		for k, v := range hdrs {
			r.SetHeader(k, v)
		}
	}
	return r
}

func (r *Request) SetQuery(k, v string) *Request {
	r.qrs.Set(k, v)
	return r
}

func (r *Request) SetQueries(queries map[string]string) *Request {
	for k, v := range queries {
		r.SetQuery(k, v)
	}
	return r
}

func (r *Request) RequestHook(hook RequestHook) *Request {
	r.requestHook = hook
	return r
}

func (r *Request) ResponseHook(hook ResponseHook) *Request {
	r.responseHook = hook
	return r
}

func (r *Request) RetryHook(hook RetryHook) *Request {
	r.retryHook = hook
	return r
}

// Exec performs the HTTP request with the given method, uri, and options.
//
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
	if !r.ok {
		return nil, r.err
	}

	// initiate trace once per request if available
	r.tracer = &TraceInfo{}
	if r.client.trace {
		r.Request = r.WithContext(r.tracer.Tracer(r.Context()))
	}

	r.URL.RawQuery = r.qrs.Encode()
	if r.requestHook != nil {
		if err := r.requestHook(r); err != nil {
			return nil, fmt.Errorf("failed to execute request hook: %w", err)
		}
	}

	// Excuetion and hooks placement
	res, err := r.client.Exec(r)
	if err != nil && r.retryHook == nil {
		return nil, err
	}
	if r.retryHook != nil {
		return r.retryHook(r, r.client)
	}
	if r.responseHook != nil {
		if err := r.responseHook(res); err != nil {
			return nil, fmt.Errorf("failed to execute response hook: %w", err)
		}
	}
	return res, err
}
