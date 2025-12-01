package httpx

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
)

// Response contains embdedd [http.Response] object so all the method of [http.Response] are
// accessible. In any case Response object doesn't close the underlying body it's callers
// responsibility to close the response body[This is done deliberately to avoid double closing and
// keeping Go semantic]
type Response struct {
	*http.Response
	TraceInfo  *TraceInfo
	Decompress func(io.Reader) (io.ReadCloser, error)
	// This set body to already read so can not be read further
	IsRead bool
}

// Success checks wether the response status code is in positive range.
func (r *Response) Success() bool {
	return r.StatusCode > 199 && r.StatusCode < 300
}

// Decode will decode given value based on [Decoder] option if none provided default will be
// [JSONDecoder]. Make sure body should be pointer to variable you're trying to decode. It will
// throw error if body is already read.
//
// NOTE: As Decode will bytes in memory avoid reading large responses.
func (r *Response) Decode(v any, opts ...func(d *DecodeOptions)) error {
	decOpts := &DecodeOptions{dec: JSONDecoder{}}
	for _, o := range opts {
		o(decOpts)
	}
	b, err := r.Bytes()
	if err != nil {
		return fmt.Errorf("httpx.Decode: %w", err)
	}
	return decOpts.dec.Decode(b, v)
}

func (r *Response) Bytes() ([]byte, error) {
	if r.IsRead {
		return nil, errors.New("httpx.Bytes: body is already read")
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("httpx.Bytes: error reading the body, err: %w", err)
	}
	r.IsRead = true
	return b, nil
}

// DecompressStream decompresses well known format such as gzip, x-gzip, deflate.  Other widely used
// format such as brotli, zstd will require you to provide Decompress function inside client which
// will automatically submitted to Response object. if br,zstd decompressor func are not provided
// the stream will return error.
func (r *Response) DecompressStream() (io.ReadCloser, error) {
	v, _, err := mime.ParseMediaType(r.Header.Get("Content-Encoding"))
	if err != nil {
		return nil, fmt.Errorf("DecompressStream: failed to find mediatype: %w", err)
	}
	switch v {
	case "gzip", "x-gzip":
		return gzip.NewReader(r.Body)
	case "deflate":
		cr, err := zlib.NewReader(r.Body)
		if err != nil {
			if !errors.Is(err, zlib.ErrHeader) {
				return nil, err
			}
			// if RFC1951 format
			return flate.NewReader(r.Body), nil
		}
		return cr, nil
	case "br", "zstd":
		if r.Decompress == nil {
			return nil, fmt.Errorf("no decompressor provided for %s", v)
		}
		return r.Decompress(r.Body)
	default:
		return nil, fmt.Errorf("incompatible content encoding: %s", v)
	}
}

// MultiReadBody Provides body which can auto reset when it hits [io.EOF] error. It will throw
// error if body is already read.
func (r *Response) MultiReadBody() (*MultiReadCloser, error) {
	b, err := r.Bytes()
	if err != nil {
		return nil, err
	}
	return &MultiReadCloser{bytes.NewReader(b)}, nil
}

// MultiReadCloser automatically reset the read buffer after reading is complete, Essentially making
// it infinite reader.
type MultiReadCloser struct {
	br *bytes.Reader
}

// Read implments [io.Reader] interface.
func (r *MultiReadCloser) Read(p []byte) (int, error) {
	n, err := r.br.Read(p)
	if err == io.EOF {
		r.br.Seek(0, io.SeekStart)
	}
	return n, err
}

func (r *MultiReadCloser) Close() error {
	return nil
}
