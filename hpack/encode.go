package hpack

import (
	huff "opal/hpack/huffman"
)

// Encoder manages the encoding of headerfields
type Encoder struct {
	dynTab *dynamicTable

	buf []byte // Current working buffer
}

// NewEncoder creates a new encoder with given table size
func NewEncoder(dynTabMaxSize uint32) *Encoder {
	dynT := newDynamicTable(dynTabMaxSize)
	return &Encoder{
		dynTab: dynT,
	}
}

// EncodeField encodes a given headerfield and returns the sequence of encoded bytes
func (e *Encoder) EncodeField(hf *HeaderField) []byte {
	e.buf = make([]byte, 0)

	// Check if header exists in the static or the dynamic table
	idx, perfectMatch := e.findHFMatch(hf)

	// If found a perfect match from the table, encode indexed header field
	if perfectMatch {
		e.encodeIndexed(idx)
	} else {
		// Check if there the hf-size does not exceed the max limit of the dynamic table
		willIndex := hf.size() <= e.dynTab.maxSize
		// if so, add to dynamic table
		if willIndex {
			e.dynTab.addEntry(hf)
		}

		if idx != 0 {
			e.encodeFieldIndexed(hf, idx, willIndex)
		} else {
			e.encodeField(hf, willIndex)
		}
	}
	return e.buf
}

func (e *Encoder) encodeIndexed(idx uint32) {
	l := len(e.buf)
	e.buf = applyIndexOrLength(e.buf, 7, idx)
	e.buf[l] |= 0x80 // Sets first bit to 1 -> 0x80 = 1xxx xxxx
}

func (e *Encoder) encodeFieldIndexed(hf *HeaderField, idx uint32, isIndexed bool) {
	l := len(e.buf)
	var n byte // Number of bits to shift => xxnn nnnn
	var mask byte
	if isIndexed {
		n = 6
		mask = 64 // 0100 0000
	} else {
		n = 4
		mask = 0
	}

	e.buf = applyIndexOrLength(e.buf, n, idx)
	e.buf[l] |= mask
	buf := encodeLitrString(e.buf, hf.Value)

	e.buf = buf
}

func (e *Encoder) encodeField(hf *HeaderField, isIndexed bool) {
	var idx byte
	if isIndexed {
		idx = 64 // 0100 0000
	} else {
		idx = 0
	}

	e.buf = append(e.buf, idx) // appends 0100 0000 or 0000 0000
	buf := encodeLitrString(e.buf, hf.Name)

	e.buf = buf
	buf = encodeLitrString(e.buf, hf.Value)

	e.buf = buf
}

// ----- HELPERS -----
func encodeLitrString(buf []byte, s string) []byte {
	// Decode string with huffman
	huffDecoded := huff.Encode([]byte(s))

	// Calculate length and death
	first := len(buf)
	length := uint32(len(huffDecoded))
	buf = applyIndexOrLength(buf, 7, length)
	buf[first] |= 0x80 // Makes sure the H is set to 1 => 0x80 = 1xxxx xxxx

	buf = append(buf, huffDecoded...)
	return buf
}

func (e *Encoder) findHFMatch(hf *HeaderField) (idx uint32, perfectMatch bool) {

	// Search through the static table
	idx, perfectMatch = findStaticHF(hf)
	if perfectMatch {
		return idx, perfectMatch
	}

	// Search through dynamic table
	dynIdx, perfectMatch := e.dynTab.findIndex(hf)
	if perfectMatch || (idx == 0 && dynIdx != 0) {
		return uint32(len(staticTableEntries)) + dynIdx, perfectMatch
	}
	return idx, perfectMatch
}

func applyIndexOrLength(buf []byte, n byte, idx uint32) []byte {
	// RFC7541 5.1:
	// If the integer value is small enough, i.e., strictly less than 2^N-1, it is encoded within the N-bit prefix.
	k := uint32((1 << n) - 1) // 2^N - 1
	if idx < k {
		return append(buf, byte(idx))
	}

	// RFC641 5.1:
	// Otherwise, all the bits of the prefix are set to 1, and the value, decreased by 2^N-1, is encoded using a list of one or more octets.
	buf = append(buf, byte(k))
	idx -= k // Decreasing value by 2^N - 1

	for ; idx >= 128; idx >>= 7 {
		buf = append(buf, byte(0x80|(idx&0x7f))) // 1xxx xxxx |
	}
	return append(buf, byte(idx))
}
