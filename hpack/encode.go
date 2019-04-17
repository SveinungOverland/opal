package hpack

import (
	huff "opal/hpack/huffman"
)

type encoder struct {
	dynTab *dynamicTable

	buf []byte // Current working buffer
}

func (e *encoder) DecodeField(hf *HeaderField) ([]byte, error) {
	e.buf = make([]byte, 0)
	var err error

	// Check if header exists in the static or the dynamic table
	idx, perfectMatch := e.findHFMatch(hf)

	// If found a perfect match from the table, encode indexed header field
	if perfectMatch {
		e.encodeIndexed(byte(idx))
	} else {
		// Check if there the hf-size does not exceed the max limit of the dynamic table
		willIndex := hf.size() <= e.dynTab.maxSize
		// if so, add to dynamic table
		if willIndex {
			e.dynTab.addEntry(hf)
		}

		if idx != 0 {
			err = e.encodeFieldIndexed(hf, byte(idx), willIndex)
		} else {
			err = e.encodeField(hf, willIndex)
		}
	}
	return e.buf, err
}

func (e *encoder) encodeIndexed(idx byte) {
	idx |= 0x80 // Sets first bit to 1 -> 0x80 = 1xxx xxxx
	e.buf = append(e.buf, idx)
}

func (e *encoder) encodeFieldIndexed(hf *HeaderField, idx byte, isIndexed bool) error {
	if isIndexed {
		idx |= 64
	} else {
		idx &= 16
	}

	e.buf = append(e.buf, idx)
	buf, err := encodeLitrString(e.buf, hf.Value)
	if err != nil {
		return err
	}
	e.buf = buf
	return nil
}

func (e *encoder) encodeField(hf *HeaderField, isIndexed bool) error {
	var idx byte
	if isIndexed {
		idx = 64
	} else {
		idx = 16
	}

	e.buf = append(e.buf, idx)
	buf, err := encodeLitrString(e.buf, hf.Name)
	if err != nil {
		return err
	}
	e.buf = buf
	buf, err = encodeLitrString(e.buf, hf.Name)
	if err != nil {
		return err
	}
	e.buf = buf
	return nil
}

// ----- HELPERS -----
func encodeLitrString(buf []byte, s string) ([]byte, error) {
	// Decode string with huffman
	huffDecoded, err := huff.Decode([]byte(s))
	if err != nil {
		return nil, err
	}

	// Calculate length and death
	length := byte(len(huffDecoded))
	length |= 0x80 // Makes sure the H is set to 1 => 0x80 = 1xxxx xxxx
	buf = append(buf, length)
	buf = append(buf, huffDecoded...)
	return buf, nil
}

func (e *encoder) findHFMatch(hf *HeaderField) (idx uint32, perfectMatch bool) {

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
