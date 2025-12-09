package hooks

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"mime"

	"github.com/jshk00/collections/httpx"
)

// ResponseHook provides a response body hook with optional automatic
// decompression and parsing into a generic type [T].
//
// If [ResponseHook.AutoParse] is true and the response Content-Type
// is JSON or XML, the body will be decoded into [ResponseHook.Body].
// Otherwise, the raw decompressed body is available via [ResponseHook.RawBody].
//
// Closing is the caller’s responsibility. The caller must close
// [ResponseHook.RawBody] (if set). If [RawBody] wraps [net/http.Response.Body],
// closing it also closes the underlying body. Close operations are
// idempotent, so it is safe to close both.
type ResponseHook[T any] struct {
	Body      T
	AutoParse bool
	// If true, attempt to transparently decompress the response body based on
	// [net/http.Response] "Content-Encoding" header. By default, [net/http]
	// will handle decompression unless "Accept-Encoding" was explicitly set.
	Decompress bool
	// Standard library doesn't support brotli and zstd decompression so if they are desired you
	// will need to provide decompressor func which will host the decompression logic for desired
	// algorithm.
	//
	// NOTE: The returned [io.ReadCloser] *must* follow Go's convention that Close is idempotent.
	// This guarantees callers can safely close both RawBody and [net/http.Response.Body] without
	// error.
	Decompressor func(io.Reader) (io.ReadCloser, error)
}

// Hook implements a response hook to process [net/http.Response] bodies.
//
// If [ResponseHook.AutoParse] is true, and the Content-Type is JSON or XML,
// the body is decoded into [ResponseHook.Body]. In this case, [RawBody] will
// already have been consumed.
//
// Caller is responsible for closing [ResponseHook.RawBody].
func (r *ResponseHook[T]) Hook(res *httpx.Response) error {
	if res == nil {
		return nil
	}

	if r.AutoParse {
		mimeType, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
		if err != nil {
			return err
		}
		switch mimeType {
		case "application/json":
			if err := json.NewDecoder(res.Body).Decode(&r.Body); err != nil {
				return err
			}
		case "text/xml", "application/xml":
			if err := xml.NewDecoder(res.Body).Decode(&r.Body); err != nil {
				return err
			}
		}
	}

	return nil
}
