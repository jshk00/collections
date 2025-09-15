package httpx

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	client *http.Client
}

func New(trace bool) *Client {
	return (&Client{
		client: &http.Client{},
	}).SetTransport(defaultTransport)
}

// SetTransport set the httptransport,
// if povided transport is nil,
// default transport will be used.
func (c *Client) SetTransport(t http.RoundTripper) *Client {
	if t != nil {
		c.client.Transport = t
	}
	return c
}

// DisableRedirect disable the redirects in http.Client.
// By default redirect are not disabled and
// follows upto configured redirects in http client.
func (c *Client) DisableRedirect() *Client {
	c.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return c
}

// SetCookieJar set cookie jar with contained cookies
// by default no cookie jar is setup
func (c *Client) SetCookieJar(jar http.CookieJar) *Client {
	c.client.Jar = jar
	return c
}

// Get is http get method
func (c *Client) Get(uri string) *HTTPOptions {
	return NewHTTPOptions().Method(http.MethodGet).URL(uri)
}

// Head is http head method follows upto 10 redirect
func (c *Client) Head(uri string) *HTTPOptions {
	return NewHTTPOptions().Method(http.MethodHead).URL(uri)
}

// Post is http post method
func (c *Client) Post(uri string, body io.Reader) *HTTPOptions {
	return NewHTTPOptions().Method(http.MethodPost).URL(uri).Body(body)
}

// Put is http put method
func (c *Client) Put(uri string, body io.Reader) *HTTPOptions {
	return NewHTTPOptions().Method(http.MethodPut).URL(uri).Body(body)
}

// Patch is http patch method
func (c *Client) Patch(uri string, body io.Reader) *HTTPOptions {
	return NewHTTPOptions().Method(http.MethodPatch).URL(uri).Body(body)
}

// Delete is http delete method
func (c *Client) Delete(uri string) *HTTPOptions {
	return NewHTTPOptions().Method(http.MethodDelete).URL(uri)
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
//     (e.g. decoding JSON, logging, validation) in the Cond function itself.
//
//   - When writing a custom retryHook, encapsulate your retry decision and
//     any post-processing logic inside the retryHook implementation.
//
// This ensures hooks remain predictable and prevents accidental multiple
// reads of the response body.
func (c *Client) Exec(ho *HTTPOptions) (*http.Response, error) {
	if ho == nil {
		ho = &HTTPOptions{}
	}

	if ho.ctx == nil {
		ho.ctx = context.Background()
	}

	// if trace is available
	// if c.trace {
	// 	ho.ctx = httptrace.WithClientTrace(ho.ctx, c.tracer)
	// }

	// initiate request with context
	req, err := http.NewRequestWithContext(ho.ctx, ho.method, ho.url, ho.body)
	if err != nil {
		return nil, err
	}

	// set all optional headers
	req.Header = ho.headers
	// set all optional queries
	req.URL.RawQuery = ho.queries.Encode()

	if ho.requestHook != nil {
		if err := ho.requestHook(req); err != nil {
			return nil, fmt.Errorf("failed to execute request hook: %w", err)
		}
	}

	res, err := c.client.Do(req)
	if ho.retryHook != nil {
		return ho.retryHook(req, res, c.client, err)
	}

	if ho.responseHook != nil {
		if err := ho.responseHook(req, res); err != nil {
			return nil, fmt.Errorf("failed to execute response hook: %w", err)
		}
	}

	return res, nil
}
