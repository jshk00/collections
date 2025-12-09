package httpx

import (
	"encoding/json"
	"encoding/xml"
)

// Decoder abstracts deserialization of raw response bytes into a user-provided value.
type Decoder interface {
	Decode(data []byte, v any) error
}

// JSONDecoder implements Decoder using [json.Unmarshal].
type JSONDecoder struct{}

// Decode unmarshals JSON data into v.
func (JSONDecoder) Decode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// XMLDecoder implements Decoder using [xml.Unmarshal].
type XMLDecoder struct{}

// Decode unmarshals XML data into v.
func (XMLDecoder) Decode(data []byte, v any) error {
	return xml.Unmarshal(data, v)
}

// DecodeOptions controls how Response.Decode deserializes the response body.
// If no options are provided, JSON decoding is used by default.
type DecodeOptions struct {
	dec Decoder
}

// WithXML sets the decoder to XMLDecoder. Overrides the default JSON decoder.
func WithXML() func(do *DecodeOptions) {
	return func(do *DecodeOptions) {
		do.dec = XMLDecoder{}
	}
}

// WithJSON sets the decoder to JSONDecoder. This is redundant unless overriding
// a previous option, because JSON is the default.
func WithJSON() func(do *DecodeOptions) {
	return func(do *DecodeOptions) {
		do.dec = JSONDecoder{}
	}
}

// WithCustom sets a user-provided Decoder implementation. This allows callers
// to handle formats other than JSON/XML or apply custom deserialization logic.
func WithCustom(dec Decoder) func(do *DecodeOptions) {
	return func(do *DecodeOptions) {
		do.dec = dec
	}
}
