package opal

import (
	"crypto/tls"
	"fmt"
	"net"
	"opal/hpack"

	"opal/frame"
	"opal/frame/types"
)

type Conn struct {
	server        *Server
	conn          net.Conn
	tlsConn       *tls.Conn
	hpack         *hpack.Context
	windowSize    uint32
	isTLS         bool
	maxConcurrent uint32
	streams       map[uint32]*Stream // map streamId to Stream instance
}

func (c *Conn) serve() {
	// start := time.Now() // Request timer

	// Initialize TLS handshake
	if c.isTLS {
		err := c.tlsConn.Handshake()
		if err != nil {
			fmt.Println(err)
			c.isTLS = false
			c.tlsConn.Close()
			return
		}
	}

	prefaceBuffer := make([]byte, 24)
	c.tlsConn.Read(prefaceBuffer)
	fmt.Println(string(prefaceBuffer))

	settingsFrame := frame.ReadFrame(c.tlsConn)
	if settingsFrame.Type != frame.SettingsType {
		// This should not happen but error should be handled
	}

	//								(SettingId to SettingValue) Setting 1 is ContextSize
	c.hpack = hpack.NewContext(settingsFrame.Payload.(*types.SettingsPayload).IDValuePair[1])

	// TODO: Change actual settings based on the frame above
	settingsResponse := &frame.Frame{
		ID:     0,
		Type:   frame.SettingsType,
		Length: 0,
		Flags: &types.SettingsFlags{
			Ack: true,
		},
	}
	// TODO: Write settingsResponse to client to acknowledge settings frame
	settingsFrameBytes := settingsResponse.ToBytes()
	c.tlsConn.Write(settingsFrameBytes)

	// Connection initiated and ready to receive header frames

	for {
		newFrame := frame.ReadFrame(c.tlsConn)

		switch newFrame.Type {
		case frame.DataType:
			// Data should always be associated with a stream
			stream, ok := c.streams[newFrame.ID]
			if !ok {
				// Error, data frame is not associated with a stream
				continue
			}
			if newFrame.Flags.(*types.DataFlags).EndStream {
				stream.state = HalfClosedRemote
			}
			if stream.data == nil {
				stream.data = newFrame.Payload.(*types.DataPayload).Data
			} else {
				stream.data = append(stream.data, newFrame.Payload.(*types.DataPayload).Data...)
			}
		case frame.HeadersType:
			// New stream
			if newFrame.ID == 0 {
				// Error, a header should always be associated with a stream
				continue
			}
			streamState := Idle
			if newFrame.Flags.(*types.HeadersFlags).EndHeaders {
				streamState = Open
			}
			if newFrame.Flags.(*types.HeadersFlags).EndStream && streamState == Open {
				streamState = HalfClosedRemote
			}
			c.streams[newFrame.ID] = &Stream{
				id:               newFrame.ID,
				state: 			  streamState,
				lastFrame:        &newFrame,
				headers:          newFrame.Payload.(*types.HeadersPayload).Fragment,
				streamDependency: newFrame.Payload.(*types.HeadersPayload).StreamDependency,
				priorityWeight:   newFrame.Payload.(*types.HeadersPayload).PriorityWeight,
			}
		case frame.PriorityType:
			stream, ok := c.streams[newFrame.ID]
			if !ok || newFrame.ID == 0 {	
				// Error, a priority frame should be .... with a stream
			}
			stream.priorityWeight = newFrame.Payload.(*types.PriorityPayload).PriorityWeight
			stream.streamDependency = newFrame.Payload.(*types.PriorityPayload).StreamDependency
		case frame.RstStreamType:

		case frame.SettingsType:
		case frame.PushPromiseType:
		case frame.PingType:
		case frame.GoAwayType:
		case frame.WindowUpdateType:
			// Update the window size
			if newFrame.ID == 0 {
				c.windowSize += newFrame.Payload.(*types.WindowUpdatePayload).WindowSizeIncrement
			}
		case frame.ContinuationType:
			// Append headerfragment
			stream, ok := c.streams[newFrame.ID]
			if !ok || newFrame.ID == 0 {
				// Error continuation should always only follow a header
			}
			stream.endHeaders = newFrame.Flags.(*types.ContinuationFlags).EndHeaders
			stream.headers = append(stream.headers, newFrame.Payload.(*types.ContinuationPayload).HeaderFragment...)
		}
	}

	// windowUpdateFrame := frame.ReadFrame(c.tlsConn)
	// fmt.Printf("Window update frame %+v\n", windowUpdateFrame)
	// fmt.Printf("Window update payload %+v\n", windowUpdateFrame.Payload.(*types.WindowUpdatePayload))

	// headersFrame := frame.ReadFrame(c.tlsConn)
	// fmt.Printf("Headers frame %+v\n", headersFrame.Flags.(*types.HeadersFlags))

	// s := &Stream{
	// 	id:      headersFrame.ID,
	// 	headers: make([]*types.HeadersPayload, 0),
	// }
	// s.headers = append(s.headers, headersFrame.Payload.(*types.HeadersPayload))

	// req, err := s.Build(c.hpack)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Printf("%+v\n", req)

	// fmt.Println(c.hpack.Decode((headersFrame.Payload.(*types.HeadersPayload).Fragment)))
}
