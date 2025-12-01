package httpx

import (
	"io"
	"net/http"
)

type Client struct {
	client *http.Client
	trace  bool
	// Decompress function for response such as br, zstd
	decompFn func(io.Reader) (io.ReadCloser, error)
}

func New() *Client {
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

func (c *Client) EnableTrace() *Client {
	c.trace = true
	return c
}

func (c *Client) SetRespDecompressor(fn func(io.Reader) (io.ReadCloser, error)) *Client {
	c.decompFn = fn
	return c
}

// DisableRedirect disable the redirects in http.Client.
// By default redirect are not disabled and
// follows upto configured redirects in http client.
func (c *Client) DisableRedirect() *Client {
	c.client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
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
func (c *Client) Get(uri string) *Request {
	return NewRequest(http.MethodGet, uri, nil)
}

// Head is http head method follows upto 10 redirect
func (c *Client) Head(uri string) *Request {
	return NewRequest(http.MethodHead, uri, nil)
}

// Post is http post method
func (c *Client) Post(uri string, body io.Reader) *Request {
	return NewRequest(http.MethodPost, uri, body)
}

// Put is http put method
func (c *Client) Put(uri string, body io.Reader) *Request {
	return NewRequest(http.MethodPut, uri, body)
}

// Patch is http patch method
func (c *Client) Patch(uri string, body io.Reader) *Request {
	return NewRequest(http.MethodPatch, uri, body)
}

// Delete is http delete method
func (c *Client) Delete(uri string) *Request {
	return NewRequest(http.MethodDelete, uri, nil)
}

func (c *Client) Exec(r *Request) (*Response, error) {
	res, err := c.client.Do(r.Request) // nolint:bodyClose
	return &Response{Response: res, TraceInfo: r.tracer, Decompress: c.decompFn}, err
}
