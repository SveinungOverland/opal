package hpack

import (
	"fmt"
)

// HeaderField contains the name and the value of a header
type HeaderField struct {
	Name  string
	Value string
}

func (hf *HeaderField) size() uint32 {
	/* RFC7541 - Section 4.1 states the following:
	The size of an entry is the sum of its name's length in octets (as defined in Section 5.2), its value's length in octets, and 32.
	The size of an entry is calculated using the length of its name and value without any Huffman encoding applied. */
	return uint32(len(hf.Name) + len(hf.Value) + 32)
}

func (hf *HeaderField) String() string {
	return fmt.Sprintf("%s: %s", hf.Name, hf.Value)
}

type dynamicTable struct {
	size         uint32 // The size of the dynamic table (sum of the size of entries)
	maxSize      uint32 // Max size
	HeaderFields []*HeaderField
}

func newDynamicTable(maxSize uint32) *dynamicTable {
	return &dynamicTable{
		size:         0,
		maxSize:      maxSize,
		HeaderFields: make([]*HeaderField, 0),
	}
}

// Checks if size is larger than max size, and if so, removes entires
func (dynT *dynamicTable) evictionCheck() bool {
	// Calculate how many entires needs to be removed
	var n uint32 // Number of entries to remove

	// As long as the size is larger than max size, remove an entry
	for dynT.size > dynT.maxSize && n < dynT.length() {
		dynT.size -= dynT.HeaderFields[dynT.length()-1-n].size()
		n++
	}
	dynT.evict(n)
	return n > 0
}

// Evicts n of the oldest entries in the table
func (dynT *dynamicTable) evict(n uint32) {
	if n == 0 {
		return
	}
	if n >= dynT.length() {
		dynT.HeaderFields = make([]*HeaderField, 0)
	}

	// Evicts n number of entries
	dynT.HeaderFields = dynT.HeaderFields[:dynT.length()-n]
}

func (dynT *dynamicTable) addEntry(hf *HeaderField) {
	/* 	RFC7541 - Section 2.3.2 states the following:
	The dynamic table consists of a list of header fields maintained in first-in, first-out order.
	The first and newest entry in a dynamic table is at the lowest index,
	and the oldest entry of a dynamic table is at the highest index. */
	dynT.size += hf.size()
	dynT.HeaderFields = append([]*HeaderField{hf}, dynT.HeaderFields...) // Add entry at index 0
	dynT.evictionCheck()
}

func (dynT *dynamicTable) length() uint32 {
	return uint32(len(dynT.HeaderFields))
}

func (dynT *dynamicTable) get(index uint32) *HeaderField {
	if index <= 0 || index > uint32(len(dynT.HeaderFields)) {
		return nil
	}
	return dynT.HeaderFields[index-1] // The index address space starts at index 1, not 0
}

func (dynT *dynamicTable) setMaxSize(size uint32) {
	dynT.maxSize = size
	dynT.evictionCheck()
}
