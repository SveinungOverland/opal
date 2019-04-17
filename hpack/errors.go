package hpack

import "fmt"

type decodingError struct {
	Err error
}

func (de decodingError) Error() string {
	return fmt.Sprintf("decoding error: %v", de.Err)
}
