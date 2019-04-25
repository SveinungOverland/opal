package opal

import (
	"fmt"
	"opal/frame"
	"opal/frame/types"
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
	queueHealthCheck := func() {
		fmt.Println("index:", index, "head:", head, "indexesLeft:", indexesLeft)
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
	}

	addFrame := func(f *frame.Frame) {
		queueHealthCheck()
		frames[index] = f
		indexesLeft--
		index++
	}
	
	addStream := func(s *Stream) {
		// fmt.Printf("Adding stream %+v\n", s)
		// TODO: Check stream state, to make sure client is waiting to receive frames
		maxPayloadSize := c.settings[5]
		// Create Header Frame
		headerLength := uint32(len(s.headers))
		headerFlags := &types.HeadersFlags{}
		if s.streamDependency != 0 {
			headerFlags.Priority = true
			headerLength += 5
		}
		headerFramesNeeded := (headerLength + maxPayloadSize - 1) / maxPayloadSize // Ceil of int division
		if headerFramesNeeded == 1 {
			headerFlags.EndHeaders = true
		}
		if len(s.data) == 0 {
			headerFlags.EndStream = true
		}
		offset := uint32(0)
		headersPayload := &types.HeadersPayload{}
		if headerFlags.Priority {
			offset = 5 // For the 5 bytes streamdependency and priorityweight uses
			headersPayload.StreamDependency = s.streamDependency
			headersPayload.PriorityWeight = s.priorityWeight
		}									// TODO: Fix this
		headersPayload.Fragment = s.headers //[:maxPayloadSize-offset] // Subtracting length in case priority flag is set 
		headerFrame := &frame.Frame{
			ID: s.id,
			Type: frame.HeadersType,
			Flags: headerFlags,
			Payload: headersPayload,
			Length: uint32(len(headersPayload.Fragment)) + offset,
		}
		addFrame(headerFrame)
		for i := uint32(1); i < headerFramesNeeded; i++ {
			// Create Continuation frames for the remaining header bytes
			headerFragment := s.headers[i*maxPayloadSize-offset:][:i*maxPayloadSize]
			flags := &types.ContinuationFlags{}
			if i == headerFramesNeeded-1 {
				flags.EndHeaders = true
			}
			continuationFrame := &frame.Frame{
				ID: s.id,
				Type: frame.ContinuationType,
				Flags: flags,
				Payload: &types.ContinuationPayload{
					HeaderFragment: headerFragment,
				},
				Length: uint32(len(headerFragment)),
			}
			addFrame(continuationFrame)
		}
		// Create payload frames
		// fmt.Println(s.data)
		dataLength := uint32(len(s.data))
		// fmt.Println("DATALENGTH:::", dataLength)
		dataFramesNeeded := (dataLength + maxPayloadSize - 1) / maxPayloadSize // Ceil of int division
		for i := uint32(0); i < dataFramesNeeded; i++ {
			dataFlags := &types.DataFlags{}
			if i == dataFramesNeeded-1 {
				dataFlags.EndStream = true
			}
			data := s.data[i*maxPayloadSize:] // [:maxPayloadSize*i]
			dataFrame := &frame.Frame{
				ID: s.id,
				Type: frame.DataType,
				Flags: dataFlags,
				Payload: &types.DataPayload{
					Data: data,
				},
				Length: uint32(len(data)),
			}
			addFrame(dataFrame)
		}
	}

	// Listen for new writable streams or frame
	for {
		// fmt.Println("StreamWriter is looping")
		queueHealthCheck()
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
			// fmt.Printf("Writing frame %+v\n with flags %+v\n and payload: %+v\n", frames[head], frames[head].Flags, frames[head].Payload)
			c.tlsConn.Write(frames[head].ToBytes())
			head++
			indexesLeft++

		}
	}
}
