package opal

import (
	"opal/frame"
)

/*
	This file contains the function that handles writing streams
	to the client through the outChan channel made for each
	connection (Conn)
*/


type OutChanWrapper struct {
	stream *Stream
	frame *frame.Frame
}


func WriteStream(c *Conn) {
	// Init variables needed for function

	// Listen for new writable streams or frame
	for {
		select {
		case stream := <- c.outChan:
		case frame := <- c.outChanFrame:
		}
	}
}
