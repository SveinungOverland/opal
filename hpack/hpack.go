package hpack

import "fmt"

type Context struct {
	dynT    *dynamicTable
	decoder *decoder
	encoder *encoder
}

// NewContext creates a new hpack-context. It initializes a new dynamic table with a given
// max-size.
func NewContext(dynamicTableMaxSize uint32) *Context {
	// Initialize dynamic table
	dynT := newDynamicTable(dynamicTableMaxSize)
	decoder := newDecoder(dynT)
	encoder := newEncoder(dynT)

	return &Context{
		dynT:    dynT,
		decoder: decoder,
		encoder: encoder,
	}
}

// Decode decodes a sequence of bytes from a header frame.
// Returns an array of HeaderFields
func (c *Context) Decode(bytes []byte) ([]*HeaderField, error) {
	return c.decoder.Decode(bytes)
}

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

// DynamicTable returns a deep copy of the HeaderFields in the dynamic table
func (c *Context) DynamicTable() []*HeaderField {
	hfs := make([]*HeaderField, len(c.dynT.HeaderFields))
	copy(hfs, c.dynT.HeaderFields)
	return hfs
}

func (c *Context) DynamicTableString() string {
	var output string
	output += fmt.Sprintf("Table size: %d\n", c.dynT.size)
	for i, hf := range c.dynT.HeaderFields {
		output += fmt.Sprintf("%d [s = %d] - %s\n", i, hf.size(), hf.String())
	}
	return output
}
