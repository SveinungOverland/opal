package hpack

import "fmt"

type DecodingError struct {
	Err error
}

func (de DecodingError) Error() string {
	return fmt.Sprintf("decoding error: %v", de.Err)
}
