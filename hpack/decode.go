package hpack

import (
	"errors"
	"fmt"
	huff "opal/hpack/huffman"
)

// Decoder manages the decoding of headerfields
type Decoder struct {
	dynTab *dynamicTable

	buf []byte // The current working buffer

	onHeaderParsed func(hf *HeaderField) // A function for handling when a headerfield is successfully parsed
}

// NewDecoder creates a new decoder with given table size
func NewDecoder(dynTabMaxSize uint32) *Decoder {
	dynT := newDynamicTable(dynTabMaxSize)
	return &Decoder{
		dynTab: dynT,
	}
}

// Decode decodes a byte array of headers
func (d *Decoder) Decode(buf []byte) ([]*HeaderField, error) {
	hfields := make([]*HeaderField, 0)
	d.onHeaderParsed = func(hf *HeaderField) { hfields = append(hfields, hf) }
	err := d.decodeHeaders(buf)
	if err != nil {
		return nil, err
	}
	return hfields, nil
}

func (d *Decoder) decodeHeaders(buf []byte) error {
	d.buf = buf

	var err error
	for len(d.buf) > 0 {

		// Read the header field representation
		hfRepr := d.getHeaderFieldRepr()
		if hfRepr == invalidHFRepr {
			return decodingError{errors.New("Invalid encoding")}
		}

		// Parse header field
		switch {
		case hfRepr == indexed:
			err = d.parseIndexedField()
		case hfRepr == litrWithIndex:
			err = d.parseLiteralString(6, litrWithIndex)
		case hfRepr == litrWithoutIndex:
			err = d.parseLiteralString(4, litrWithoutIndex)
		case hfRepr == litrNeverIndexed:
			err = d.parseLiteralString(4, litrNeverIndexed)
		case hfRepr == dynTabSizeUpdate:
			err = d.parseDynTabSizeUpdate()
		}
		if err != nil {
			break
		}
	}

	return err
}

func (d *Decoder) getHeaderFieldRepr() headerFieldRepr {
	b := d.buf[0]
	switch {
	// Indexed representation - MSB are set to 1 - Section 6.2.1
	case b&128 != 0:
		return indexed

	// Literal Header Field with Incremental Indexing - Starts with '01'xx xxxx - Section 6.2.2
	case b&192 == 64:
		return litrWithIndex

	// Literal Header Field without Indexing - Starts with '0000' xxxx - Section 6.2.3
	case b&240 == 0:
		return litrWithoutIndex

	// Literal Header Field never Indexed - Starts with '0001' xxxx - Section 6.2.4
	case b&240 == 16:
		return litrNeverIndexed

	// Update Size of Dynamic Table - Starts with '001'x xxxx - Section 6.3
	case b&224 == 32:
		return dynTabSizeUpdate
	}

	// Decode error
	return invalidHFRepr
}

func (d *Decoder) getHeaderFieldByIndex(index uint32) (*HeaderField, bool) {

	// Indexes starts at 1, not 0. Index can not be longer than the static table
	if index == 0 {
		return nil, false
	}

	// Check if index satistifes the static table
	if index < uint32(len(staticTableEntries)) {
		return getStaticHF(index), true
	}

	// Check if index is greater than the dynamic table
	if index > uint32(len(staticTableEntries))+d.dynTab.length() {
		return nil, false
	}

	// Get HeaderField from dynamic table
	// Dynamic table indexes starts at 1
	hf := d.dynTab.get(index - uint32(len(staticTableEntries)))
	if hf == nil {
		return hf, false
	}
	return hf, true

}

// Section 6.1 - http://http2.github.io/http2-spec/compression.html#indexed.header.representation
func (d *Decoder) parseIndexedField() error {
	idx, buf, err := readLSBValue(7, d.buf) // Gets the 7 LSB - 0xxx xxxx
	if err != nil {
		return err
	}

	hf, ok := d.getHeaderFieldByIndex(idx)
	if !ok {
		return decodingError{errors.New(fmt.Sprintf("Invalid index: %d", idx))}
	}

	d.buf = buf
	d.onHeaderParsed(hf) // Successfully parsed hf
	return nil
}

// Parses an literal string
// Section 6.2.1 - http://http2.github.io/http2-spec/compression.html#indexed.header.representation
func (d *Decoder) parseLiteralString(n byte, hfRepr headerFieldRepr) error {
	buf := d.buf
	idx, buf, err := readLSBValue(n, buf) // Gets the n LSB - xxnn nnnn
	if err != nil {
		return err
	}

	// --- Get Name of header ---
	hf := &HeaderField{}

	// If idx is not 0 get name from table
	if idx > 0 {
		hf2, ok := d.getHeaderFieldByIndex(uint32(idx))
		if !ok {
			return decodingError{errors.New(fmt.Sprintf("Invalid index: %d", idx))}
		}
		hf.Name = hf2.Name
	} else {
		// If idx is 0, get name from string value
		name, remainBuf, err := readLiteralString(buf)
		if err != nil {
			return err
		}
		hf.Name = name
		buf = remainBuf
	}

	// ---- Read header value ----
	value, buf, err := readLiteralString(buf)
	if err != nil {
		return err
	}
	hf.Value = value
	d.buf = buf // Removes read bytes

	// Section 6.2.2 says if "With Incremental Index" should be added to the dynamic table
	if hfRepr == litrWithIndex {
		d.dynTab.addEntry(hf)
	}

	d.onHeaderParsed(hf) // Successfully parsed hf
	return nil
}

/* 0   1   2   3   4   5   6   7
+---+---+---+---+---+---+---+---+
| 0 | 0 | 1 |   Max size (5+)   |
+---+---------------------------+ */

// Sets the max size of the dynamic table
func (d *Decoder) parseDynTabSizeUpdate() error {
	// Might have to check if allowed to change dynamic table size? RFC7541 - Section 4.2

	// Read new max size
	buf := d.buf
	size, buf, err := readLSBValue(5, buf)
	if err != nil {
		return err
	}

	// Set max size
	d.dynTab.setMaxSize(size)
	d.buf = buf
	return nil
}

// ---- HELPERS ------

/* 0   1   2   3   4   5   6   7
+---+---+---+---+---+---+---+---+
| H |    String Length (7+)     |
+---+---------------------------+
|  String Data (Length octets)  |
+-------------------------------+ */

// Reads a Literal String - Section 5.2 - http://http2.github.io/http2-spec/compression.html#string.literal.representation
func readLiteralString(buf []byte) (s string, remaining []byte, err error) {
	if len(buf) == 0 {
		return "", nil, errors.New("Need more data to read")
	}

	// Read huffman value and length
	isHuffman := buf[0]&128 != 0                   // H
	stringLength, buf, err := readLSBValue(7, buf) // Gets the 7 LSB - 0xxx xxxx
	if err != nil {
		return "", nil, err
	}

	// Read string value
	// If is not huffman encoded, return the bytes in form of a string
	if !isHuffman {
		return string(buf[:stringLength]), buf[stringLength:], nil
	}

	// If huffman encoded, decode string
	decoded, err := huff.Decode(buf[:stringLength])
	if err != nil {
		return "", nil, err
	}

	return string(decoded), buf[stringLength:], err
}

/*
0   1   2   3   4   5   6   7
+---+---+---+---+---+---+---+---+
| ? | ? | ? | 1   1   1   1   1 |
+---+---+---+-------------------+
| 1 |    Value-(2^N-1) LSB      |
+---+---------------------------+
               ...
+---+---------------------------+
| 0 |    Value-(2^N-1) MSB      |
+---+---------------------------+ */

// Reads the least significant bits
func readLSBValue(n byte, buf []byte) (uint32, []byte, error) {
	if len(buf) == 0 {
		return 0, buf, errors.New("Need more data to read")
	}

	var value uint32
	var err error
	switch {
	case n == 7:
		value = uint32(buf[0] & 127) // 127 = 0111 1111
	case n == 6:
		value = uint32(buf[0] & 63) // 63 = 0011 1111
	case n == 5:
		value = uint32(buf[0] & 31) // 31 = 0001 1111
	case n == 4:
		value = uint32(buf[0] & 15) // 15 = 0000 1111
	default:
		err = errors.New("Invalid LSB n")
	}

	// Check if value is less than 2^N - 1, if so, return that value
	// Pseudo code from RFC7541:
	// if I < 2^N - 1, return I
	if value < (1<<uint32(n) - 1) {
		return value, buf[1:], err
	}

	// Otherwise, the value is actually longer. Have to read bytes until
	// the MSB of the next byte is 0. Read RFC7541 - Section 5.1

	// Pseudo code from RFC7541 - Section 5.1:
	/* decode I from the next N bits
	else
		M = 0
		repeat
			B = next octet
			I = I + (B & 127) * 2^M
			M = M + 7
		while B & 128 == 128
		return I
	*/
	var m uint32
	tempBuf := buf[1:]
	for len(tempBuf) > 0 {
		b := tempBuf[0]
		tempBuf = tempBuf[1:]
		value += uint32(b&127) << m //  I + (B & 127) * 2^M

		// Check if MSB is 0, then it is done
		if b&128 == 0 {
			return value, tempBuf, nil
		}
		m += 7
	}

	return 0, tempBuf, errors.New("Need more data to read")
}
