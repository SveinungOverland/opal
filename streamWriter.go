package opal

import (
	"container/list"
	"opal/frame"
	"opal/frame/types"
)

/*
	This file contains the function that handles writing streams
	to the client through the outChan channel made for each
	connection (Conn)
*/

func WriteStream(c *Conn) {
	// Init variables needed for function
	frames := list.New() // Queue for frames

	// Helper funcs
	addFrame := func(f *frame.Frame) {
		if f == nil {
			return
		}
		frames.PushBack(f)
	}

	addStream := func(s *Stream) {
		if s == nil {
			return
		}

		// TODO: Check stream state, to make sure client is waiting to receive frames
		maxPayloadSize := c.settings[5]
		// Create Header Frame
		headerLength := uint32(len(s.headers))
		headerFlags := &types.HeadersFlags{}
		headersPayload := &types.HeadersPayload{}

		offset := uint32(0) // offset is used to add space for streamdependency and priorityweight
		if s.streamDependency != 0 {
			headerFlags.Priority = true
			offset = 5 // For the 5 bytes streamdependency and priorityweight uses
			headersPayload.StreamDependency = s.streamDependency
			headersPayload.PriorityWeight = s.priorityWeight
		}

		headerFramesNeeded := ((headerLength + offset) + maxPayloadSize - 1) / maxPayloadSize // Ceil of int division

		if headerFramesNeeded == 1 {
			headerFlags.EndHeaders = true
			headersPayload.Fragment = s.headers
		} else {
			headersPayload.Fragment = s.headers[:maxPayloadSize-offset] // Subtracting offset in case priority flag is set
		}

		if len(s.data) == 0 {
			headerFlags.EndStream = true
		}

		headerFrame := &frame.Frame{
			ID:      s.id,
			Type:    frame.HeadersType,
			Flags:   headerFlags,
			Payload: headersPayload,
			Length:  uint32(len(headersPayload.Fragment)) + offset,
		}
		addFrame(headerFrame)
		for i := uint32(1); i < headerFramesNeeded; i++ {
			// Create Continuation frames for the remaining header bytes
			var headerFragment []byte
			flags := &types.ContinuationFlags{}
			if i == headerFramesNeeded-1 {
				flags.EndHeaders = true
				headerFragment = s.headers[i*maxPayloadSize-offset:]
			} else {
				headerFragment = s.headers[i*maxPayloadSize-offset:][:maxPayloadSize]
			}
			continuationFrame := &frame.Frame{
				ID:    s.id,
				Type:  frame.ContinuationType,
				Flags: flags,
				Payload: &types.ContinuationPayload{
					HeaderFragment: headerFragment,
				},
				Length: uint32(len(headerFragment)),
			}
			addFrame(continuationFrame)
		}
		// Create payload frames
		dataLength := uint32(len(s.data))
		dataFramesNeeded := (dataLength + maxPayloadSize - 1) / maxPayloadSize // Ceil of int division
		for i := uint32(0); i < dataFramesNeeded; i++ {

			var data []byte
			dataFlags := &types.DataFlags{}

			if i == dataFramesNeeded-1 {
				dataFlags.EndStream = true
				data = s.data[i*maxPayloadSize:]
			} else {
				data = s.data[i*maxPayloadSize:][:maxPayloadSize]
			}

			dataFrame := &frame.Frame{
				ID:    s.id,
				Type:  frame.DataType,
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
		select {
		// This select block is not blocking to make sure the function keeps
		// doing work if it exists
		case <-c.ctx.Done():
			return
		case stream := <-c.outChan:
			addStream(stream)
		case frame := <-c.outChanFrame:
			addFrame(frame)
		default:
			if frames.Len() == 0 {
				// If frame slice is empty reset the size to avoid memory leak
				select {
				// This select block is blocking, so this function doesn't use up
				// resources endlessly looping
				case <-c.ctx.Done():
					return
				case stream := <-c.outChan:
					addStream(stream)
				case frame := <-c.outChanFrame:
					addFrame(frame)
				}
			}
			// Write next frame
			// fmt.Printf("Writing frame %+v\n with flags %+v\n and payload: %+v\n", frames[head], frames[head].Flags, frames[head].Payload)
			frameToWrite := frames.Front()
			if frameToWrite == nil {
				return
			}
			frameValue := frameToWrite.Value
			if frameValue == nil {
				return
			}

			c.tlsConn.Write(frameValue.(*frame.Frame).ToBytes())
			frames.Remove(frameToWrite)
		}
	}
}
