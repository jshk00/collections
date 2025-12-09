package httpx

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"io"
	"sync"
)

// decompressors is concurrent safe map of decompression function.
// It already use gzip, delfate and zlib user can override it as well.
type decompressors struct {
	mu   sync.RWMutex
	data map[string]func(io.Reader) (io.ReadCloser, error)
}

func newDecompressor() *decompressors {
	return &decompressors{
		data: map[string]func(io.Reader) (io.ReadCloser, error){
			"gzip":    defaultGzipDecompressor,
			"deflate": defaultFlateDecompressor,
			"zlib":    zlib.NewReader,
		},
	}
}

func (ds *decompressors) put(key string, fn func(io.Reader) (io.ReadCloser, error)) {
	ds.mu.Lock()
	ds.data[key] = fn
	ds.mu.Unlock()
}

func (ds *decompressors) get(key string) (fn func(io.Reader) (io.ReadCloser, error), ok bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	fn, ok = ds.data[key]
	return fn, ok
}

func defaultGzipDecompressor(r io.Reader) (io.ReadCloser, error) {
	return gzip.NewReader(r)
}

func defaultFlateDecompressor(r io.Reader) (io.ReadCloser, error) {
	return flate.NewReader(r), nil
}
