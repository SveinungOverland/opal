package hpack

import "fmt"

// ----- TYPES -----

type headerFieldRepr byte

var indexed = headerFieldRepr(0)
var litrWithIndex = headerFieldRepr(1)
var litrWithoutIndex = headerFieldRepr(2)
var litrNeverIndexed = headerFieldRepr(3)
var dynTabSizeUpdate = headerFieldRepr(4)
var invalidHFRepr = headerFieldRepr(5)

// ---- ERRORS ------
type decodingError struct {
	Err error
}

func (de decodingError) Error() string {
	return fmt.Sprintf("[decoding error]: %v", de.Err)
}
