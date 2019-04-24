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
	indexesLeft := 50
	index := 0
	head := 0
	frames := make([]*frame.Frame, 50)

	// Helper funcs
	addFrame := func(f *frame.Frame) {
		frames[index] = f
		indexesLeft--
		index++
	}

	addStream := func(s *Stream) {
		
	}

	// Listen for new writable streams or frame
	for {
		if index == len(frames) {
			// Index has reached end of frames slice
			if indexesLeft != 0 {
				// frames slice has unused indexes
				index = 0
			} else {
				// indexesLeft is 0
				// frames slice has run out of space, create more
				// TODO: Release unused space from frames slice, this is a memory leak in the making
				frames = append(frames, make([]*frame.Frame, 50)...)
				indexesLeft = 50
			}
		}
		if head == len(frames) {
			if indexesLeft != 0 {
				head = 0
			}
		}
		select {
		// This select block is not blocking to make sure the function keeps
		// doing work if it exists
		case stream := <- c.outChan:
			addStream(stream)
		case frame := <- c.outChanFrame:
			addFrame(frame)
		default:
			if indexesLeft == len(frames) {
				// If frame slice is empty reset the size to avoid memory leak
				frames = make([]*frame.Frame, 50)
				select {
				// This select block is blocking, so this function doesn't use up 
				// resources endlessly looping
				case stream := <- c.outChan:
					addStream(stream)
				case frame := <- c.outChanFrame:
					addFrame(frame)
				}
			}
			// Write next frame
			c.tlsConn.Write(frames[head].ToBytes())
			head++
			indexesLeft++

		}
	}
}
