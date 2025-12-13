package httpx

import (
	"fmt"
	"io"
	"net/http"
	"sync"
)

type (
	ResponseHook     func(*Client, *Response) error
	RequestHook      func(*Client, *Request) error
	ContentTypeEncFn func(body any) (io.Reader, error)
	ContentTypeDecFn func(body any, r io.Reader) error
)

type contentTypeEncoders struct {
	mu  sync.RWMutex
	enc map[string]ContentTypeEncFn
}

func newContentTypeEncoders() *contentTypeEncoders {
	return &contentTypeEncoders{enc: make(map[string]ContentTypeEncFn)}
}

func (ce *contentTypeEncoders) set(key string, fn ContentTypeEncFn) {
	ce.mu.Lock()
	ce.enc[key] = fn
	ce.mu.Unlock()
}

func (ce *contentTypeEncoders) get(key string) (ContentTypeEncFn, bool) {
	ce.mu.RLock()
	fn, ok := ce.enc[key]
	ce.mu.RUnlock()
	return fn, ok
}

type contentTypeDecoders struct {
	mu  sync.RWMutex
	dec map[string]ContentTypeDecFn
}

func newContentTypeDecoders() *contentTypeDecoders {
	return &contentTypeDecoders{dec: make(map[string]ContentTypeDecFn)}
}

func (ce *contentTypeDecoders) set(key string, fn ContentTypeDecFn) {
	ce.mu.Lock()
	ce.dec[key] = fn
	ce.mu.Unlock()
}

func (ce *contentTypeDecoders) get(key string) (ContentTypeDecFn, bool) {
	ce.mu.RLock()
	fn, ok := ce.dec[key]
	ce.mu.RUnlock()
	return fn, ok
}

type Client struct {
	breaker             *CircuitBreaker
	client              *http.Client
	trace               bool
	decompressors       *decompressors
	contentTypeEncoders *contentTypeEncoders
	contentTypeDecoders *contentTypeDecoders
}

func New() *Client {
	return (&Client{
		client:              &http.Client{},
		decompressors:       newDecompressor(),
		contentTypeEncoders: newContentTypeEncoders(),
		contentTypeDecoders: newContentTypeDecoders(),
	}).SetTransport(defaultTransport)
}

func (c *Client) SetCircuitBreaker(b *CircuitBreaker) *Client {
	c.breaker = b
	return c
}

// SetTransport set the httptransport, if provided transport is nil, default transport will be used.
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

// DisableRedirect disable the redirects in http.Client. By default redirect are not disabled and
// follows upto configured redirects in http client.
func (c *Client) DisableRedirect() *Client {
	c.client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return c
}

// SetCookieJar set cookie jar with contained cookies by default no cookie jar is setup
func (c *Client) SetCookieJar(jar http.CookieJar) *Client {
	c.client.Jar = jar
	return c
}

// SetDecompressor registers a decompression function for the given Content-Encoding name. Keys must
// match the value of the Content-Encoding header exactly after trimming spaces.
//
// The default client provides decompressors for "gzip", "deflate", and "zlib". Calling
// SetDecompressor with an existing key overrides the default implementation.
//
// Multi-encoding responses (e.g. "gzip, zlib") are treated as a single logical encoding. The
// library does not attempt to chain multiple encodings internally. If a server sends multiple
// encodings, register a decompressor using the exact header value (e.g. "gzip, zlib") and implement
// the decoding chain inside the provided function in reverse application order:
//
//	// Example for: Content-Encoding: gzip, zlib
//	func(r io.Reader) (io.ReadCloser, error) {
//	    zr, err := zlib.NewReader(r)   // decode last applied encoding first
//	    if err != nil {
//	        return nil, err
//	    }
//	    gr, err := gzip.NewReader(zr)
//	    if err != nil {
//	        return nil, err
//	    }
//	    return gr, nil
//	}
//
// Call SetDecompressor multiple times to register additional encodings.
func (c *Client) SetDecompressor(key string, fn func(io.Reader) (io.ReadCloser, error)) *Client {
	c.decompressors.put(key, fn)
	return c
}

func (c *Client) SetContentTypeEncoder(key string, fn ContentTypeEncFn) *Client {
	c.contentTypeEncoders.set(key, fn)
	return c
}

func (c *Client) GetContentTypeEncoder(key string) (ContentTypeEncFn, bool) {
	return c.contentTypeEncoders.get(key)
}

func (c *Client) SetContentTypeDecoder(key string, fn ContentTypeDecFn) *Client {
	c.contentTypeDecoders.set(key, fn)
	return c
}

func (c *Client) GetContentTypeDecoder(key string) (ContentTypeDecFn, bool) {
	return c.contentTypeDecoders.get(key)
}

// Get is http get method
func (c *Client) Get(url string) *Request {
	return NewRequest().SetMethod(http.MethodGet).SetURL(url)
}

// Head is http head method follows upto 10 redirect
func (c *Client) Head(url string) *Request {
	return NewRequest().SetMethod(http.MethodHead).SetURL(url)
}

// Post is http post method
func (c *Client) Post(url string, body any) *Request {
	return NewRequest().SetMethod(http.MethodPost).SetURL(url).SetBody(body)
}

// Put is http put method
func (c *Client) Put(url string, body any) *Request {
	return NewRequest().SetMethod(http.MethodPut).SetURL(url).SetBody(body)
}

// Patch is http patch method
func (c *Client) Patch(url string, body any) *Request {
	return NewRequest().SetMethod(http.MethodPost).SetURL(url).SetBody(body)
}

// Delete is http delete method
func (c *Client) Delete(url string) *Request {
	return NewRequest().SetMethod(http.MethodDelete).SetURL(url)
}

func (c *Client) exec(r *Request) (*Response, error) {
	// Execute all the request hooks
	for i := 0; i < len(r.requestHook); i++ {
		if err := r.requestHook[i](c, r); err != nil {
			return nil, fmt.Errorf("failed to execute request hook: %w", err)
		}
	}

	res, err := c.client.Do(r.RawRequest) // nolint:bodyClose
	return &Response{
		Response:            res,
		traceInfo:           r.tracer,
		decompressors:       c.decompressors,
		contentTypeDecoders: c.contentTypeDecoders,
	}, err
}
