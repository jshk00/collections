package httpx

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime"
	"strings"
)

// FIXME: Check if r.Body is not written prematurely
func DefaultRequestHook(c *Client, r *Request) error {
	if r.Body != nil {
		rr, err := handleRequestBody(c, r)
		if err != nil {
			return err
		}
		r.Body = rr
	}
	return nil
}

func handleRequestBody(c *Client, r *Request) (io.Reader, error) {
	if strings.TrimSpace(r.Header.Get("Content-Type")) == "" {
		return nil, errors.New("empty content type cannot encode the body")
	}
	switch v := r.Body.(type) {
	case io.Reader:
		return v, nil
	case string:
		return strings.NewReader(v), nil
	case []byte:
		return bytes.NewReader(v), nil
	default:
		mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			return nil, err
		}
		if mt == "application/json" {
			b, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			return bytes.NewReader(b), nil
		}
		if mt == "application/xml" {
			b, err := xml.Marshal(v)
			if err != nil {
				return nil, err
			}
			return bytes.NewReader(b), nil
		}
		enc, ok := c.GetContentTypeEncoder(mt)
		if !ok {
			return nil, fmt.Errorf("content type encoder is not found for content type %s", mt)
		}
		return enc(v)
	}
}
