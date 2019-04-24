package hpack

import (
	"fmt"
	"strings"
)

const defaultDynTabSize = 4096

type Context struct {
	decoder *decoder
	encoder *encoder
}

// NewContext creates a new hpack-context. It initializes a new dynamic table with a given
// max-size.
func NewContext(dynamicTableMaxSize uint32) *Context {
	
	// Initialize Decoder
	dynT := newDynamicTable(dynamicTableMaxSize)
	decoder := newDecoder(dynT)

	// Initialize Encoder
	encodeDynt := newDynamicTable(dynamicTableMaxSize)
	encoder := newEncoder(encodeDynt)

	return &Context{
		decoder: decoder,
		encoder: encoder,
	}
}

// Decode decodes a sequence of bytes from a header frame.
// Returns an array of HeaderFields
func (c *Context) Decode(bytes []byte) ([]*HeaderField, error) {
	return c.decoder.Decode(bytes)
}

// Encode encodes a set of headers
func (c *Context) Encode(hfs []*HeaderField) ([]byte, error) {
	var bytes []byte
	for _, hf := range hfs {
		buf, err := c.encoder.EncodeField(hf)
		if err != nil {
			return bytes, err
		}
		bytes = append(bytes, buf...)
	}
	return bytes, nil
}

// EncodeHeaders encodes a set of headers from a map
func (c *Context) EncodeMap(headers map[string]string) ([]byte, error) {
	var bytes []byte
	for k, v := range headers {
		buf, err := c.encoder.EncodeField(&HeaderField{Name: strings.ToLower(k), Value: v})
		if err != nil {
			return bytes, err
		}
		bytes = append(bytes, buf...)
	}
	return bytes, nil
}

// DecoderDynamicTable returns a deep copy of the HeaderFields in the decoder's dynamic table
func (c *Context) DecoderDynamicTable() []*HeaderField {
	hfs := make([]*HeaderField, len(c.decoder.dynTab.HeaderFields))
	copy(hfs, c.decoder.dynTab.HeaderFields)
	return hfs
}

// EncoderDynamicTable returns a deep copy of the HeaderFields in the encoder's dynamic table
func (c *Context) EncoderDynamicTable() []*HeaderField {
	hfs := make([]*HeaderField, len(c.encoder.dynTab.HeaderFields))
	copy(hfs, c.encoder.dynTab.HeaderFields)
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
