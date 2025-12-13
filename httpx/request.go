package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Request struct {
	RawRequest   *http.Request
	responseHook ResponseHook
	requestHook  []RequestHook
	client       *Client
	tracer       *TraceInfo
	ctx          context.Context
	cookie       *http.Cookie
	IsTrace      bool
	URI          string
	Queries      url.Values
	Header       http.Header
	Body         any
	Method       string
	retry        *Retry
	IsRetry      bool
	Attempt      int
}

func NewRequest() *Request {
	return &Request{
		Header:      make(http.Header),
		Queries:     make(url.Values),
		requestHook: []RequestHook{DefaultRequestHook},
	}
}

func (r *Request) WithContext(ctx context.Context) *Request {
	r.ctx = ctx
	return r
}

func (r *Request) Context() context.Context {
	return r.ctx
}

func (r *Request) SetMethod(v string) *Request {
	r.Method = v
	return r
}

func (r *Request) EnableTrace() *Request {
	r.IsTrace = true
	return r
}

func (r *Request) SetRetry(retry *Retry) *Request {
	if retry == nil {
		retry = NewRetry()
	}
	r.retry = retry
	r.IsRetry = true
	return r
}

func (r *Request) SetBody(v any) *Request {
	r.Body = v
	return r
}

func (r *Request) SetURL(uri string) *Request {
	r.URI = uri
	return r
}

func (r *Request) URL() string {
	return r.URI
}

func (r *Request) SetHeader(k, v string) *Request {
	r.Header.Set(k, v)
	return r
}

func (r *Request) SetCookies(c *http.Cookie) *Request {
	r.cookie = c
	return r
}

func (r *Request) SetHeaders(hdrs map[string]string) *Request {
	for k, v := range hdrs {
		r.SetHeader(k, v)
	}
	return r
}

func (r *Request) SetQuery(k, v string) *Request {
	r.Queries.Set(k, v)
	return r
}

func (r *Request) SetQueries(queries map[string]string) *Request {
	for k, v := range queries {
		r.SetQuery(k, v)
	}
	return r
}

func (r *Request) SetRequestHook(hook RequestHook) *Request {
	r.requestHook = append(r.requestHook, hook)
	return r
}

func (r *Request) SetResponseHook(hook ResponseHook) *Request {
	r.responseHook = hook
	return r
}

// Hook execution order:
//
//  1. requestHook — runs before sending the request.
//  2. retry — if enabled, takes full control over retries and
//     determines the final response. In this case, responseHook is
//     NOT invoked.
//  3. responseHook — runs only if no retryHook is defined.
//
// Important:
//
//   - If retry is set (custom or default), responseHook will be ignored.
//     This avoids conflicts from reading res.Body multiple times.
//
//   - When using the default retry, place any post-processing logic
//     (e.g. decoding JSON, logging, validation) in the Cond function itself.
func (r *Request) Exec() (*Response, error) {
	var (
		totalWait time.Duration
		err       error
	)

	if r.IsRetry {
		for attempt := 1; attempt <= r.retry.PollLimit; attempt++ {
			r.Attempt = attempt
			res, err := r.client.exec(r)
			if err != nil {
				ctxErr := r.Context().Err()
				if ctxErr != nil && errors.Is(ctxErr, context.DeadlineExceeded) {
					return nil, ctxErr
				}
			}
			if !r.retry.Cond(res, err) {
				return res, nil
			}
			// drain some of the resposne body before wait so tcp keep alive be reuse the connection
			if res != nil && res.Body != nil {
				_, _ = io.CopyN(io.Discard, res.Body, 2048)
				res.Body.Close()
			}
			if r.retry.Backoff != nil {
				r.retry.Wait = r.retry.Backoff.NextWaitDuration(res, attempt)
			}
			totalWait += r.retry.Wait
			time.Sleep(r.retry.Wait)
		}
		return nil, RetryPollError{
			Attempts:       r.retry.PollLimit,
			TotalSleepTime: totalWait,
			ReqURL:         r.URI,
			ReqMethod:      r.Method,
			ResponseError:  err,
		}
	}

	res, err := r.client.exec(r)
	if r.responseHook != nil && r.requestHook == nil {
		if err := r.responseHook(r.client, res); err != nil {
			return nil, fmt.Errorf("failed to execute response hook: %w", err)
		}
	}
	return res, err
}
