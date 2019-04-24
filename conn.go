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
	streams       map[uint32]Stream // map streamId to Stream instance
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
		case frame.HeadersType:
			// New stream
			if newFrame.ID == 0 {
				// Error, a header should always be associated with a stream
			}
			c.streams[newFrame.ID] = Stream{
				id:               newFrame.ID,
				headers:          newFrame.Payload.(*types.HeadersPayload).Fragment,
				lastFrame:        &newFrame,
				streamDependency: newFrame.Payload.(*types.HeadersPayload).StreamDependency,
				priorityWeight:   newFrame.Payload.(*types.HeadersPayload).PriorityWeight,
			}
		case frame.PriorityType:
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
