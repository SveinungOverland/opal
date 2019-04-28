package hpack

import (
	"fmt"
)

const defaultDynTabSize = 4096

// Context manages both HPACK encoding and decoding. Is basically a wrapper.
type Context struct {
	Decoder *Decoder
	Encoder *Encoder
}

// NewContext creates a new hpack-context. It initializes a new dynamic table with a given
// max-size.
func NewContext(encoderDynTabMaxSize uint32, decoderDynTabMaxSize uint32) *Context {
	
	// Initialize Decoder
	decoder := NewDecoder(decoderDynTabMaxSize)

	// Initialize Encoder
	encoder := NewEncoder(encoderDynTabMaxSize)

	return &Context{
		Decoder: decoder,
		Encoder: encoder,
	}
}

// Decode decodes a sequence of bytes from a header frame.
// Returns an array of HeaderFields
func (c *Context) Decode(bytes []byte) ([]*HeaderField, error) {
	return c.Decoder.Decode(bytes)
}

// Encode encodes a set of headers
func (c *Context) Encode(hfs []*HeaderField) ([]byte) {
	var bytes []byte
	for _, hf := range hfs {
		buf := c.Encoder.EncodeField(hf)
		bytes = append(bytes, buf...)
	}
	return bytes
}

// DecoderDynamicTable returns a deep copy of the HeaderFields in the decoder's dynamic table
func (c *Context) DecoderDynamicTable() []*HeaderField {
	hfs := make([]*HeaderField, len(c.Decoder.dynTab.HeaderFields))
	copy(hfs, c.Decoder.dynTab.HeaderFields)
	return hfs
}

// EncoderDynamicTable returns a deep copy of the HeaderFields in the encoder's dynamic table
func (c *Context) EncoderDynamicTable() []*HeaderField {
	hfs := make([]*HeaderField, len(c.Encoder.dynTab.HeaderFields))
	copy(hfs, c.Encoder.dynTab.HeaderFields)
	return hfs
}

// dynamicTableString returns a string visualizing the state of a dynamic table
func (c *Context) dynamicTableString(dynT dynamicTable) string {
	var output string
	output += fmt.Sprintf("Table size: %d\n", dynT.size)
	for i, hf := range dynT.HeaderFields {
		output += fmt.Sprintf("%d [s = %d] - %s\n", i, hf.size(), hf.String())
	}
	return output
}
