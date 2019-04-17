package hpack

import "fmt"

type Context struct {
	dynT    *dynamicTable
	decoder *decoder
}

// NewContext creates a new hpack-context. It initializes a new dynamic table with a given
// max-size.
func NewContext(dynamicTableMaxSize uint32) *Context {
	// Initialize dynamic table
	dynT := newDynamicTable(dynamicTableMaxSize)
	decoder := newDecoder(dynT)

	return &Context{
		dynT:    dynT,
		decoder: decoder,
	}
}

// Decode decodes a sequence of bytes from a header frame.
// Returns an array of HeaderFields
func (c *Context) Decode(bytes []byte) ([]*HeaderField, error) {
	return c.decoder.Decode(bytes)
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
