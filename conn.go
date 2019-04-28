package opal

import (
	"crypto/tls"
	"fmt"
	"net"
	"opal/hpack"

	"opal/frame"
	"opal/frame/types"
	"opal/errors"

	"context"
)

const initialHeaderTableSize = uint32(4096)

type Conn struct {
	ctx           context.Context
	cancel        context.CancelFunc
	server        *Server
	conn          net.Conn
	tlsConn       *tls.Conn
	hpack         *hpack.Context
	lastReceivedFrame *frame.Frame
	windowSize    uint32
	isTLS         bool
	maxConcurrent uint32
	streams       map[uint32]*Stream // map streamId to Stream instance
	inChan		  chan *Stream // Channel for handling new ended stream
	outChan       chan *Stream // Channel for sending finished streams
	outChanFrame  chan *frame.Frame // Channel for sending single Frame's
	settings      map[uint16]uint32
	prevStreamID  uint32 // The previous created stream's identifer.
}

func (c *Conn) serve() {
	// start := time.Now() // Request timer
	defer c.cancel()
	defer close(c.inChan)
	defer close(c.outChan)
	defer close(c.outChanFrame)

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

	settingsFrame, err := frame.ReadFrame(c.tlsConn)
	if err != nil {
		// TODO: Handle error
	}
	fmt.Printf("HANDSHAKE FRAME: %+v\n", settingsFrame)
	if settingsFrame.Type != frame.SettingsType {
		// This should not happen but error should be handled
		panic("Settings frame from handshake is of wrong type!")
	}

	if settingsFrame.Length > 0 {
		for key, value := range settingsFrame.Payload.(*types.SettingsPayload).IDValuePair {
			if key >= 0x1 && key <= 0x6 {
				// Any other key is out of range and is ignored
				c.settings[key] = value
			}
		}
	}

	// Creating new HPACK context (with encoder and decoder)
	// Setting 1 is ContextSize
	c.hpack = hpack.NewContext(initialHeaderTableSize, c.settings[1])

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
	c.outChanFrame <- settingsResponse
	
	go serveStreamHandler(c) // Starting go-routine that is responsible for handling requests when streams are done
	go WriteStream(c) // Starting go-routine that is responsible for handling handled requests that should be written back to client
	
	// Connection initiated and ready to receive header frames
	// errors.EnhanceYourCalm
	for {
		// fmt.Println("Looping serve")
		newFrame, err := frame.ReadFrame(c.tlsConn)
		if err != nil {
			break
		}

		switch newFrame.Type {
		case frame.DataType:
			// Data should always be associated with a stream
			stream, ok := c.streams[newFrame.ID]
			if stream.id == 0 {
				goaway := &frame.Frame{
					ID: 0,
					Type: frame.GoAwayType,
					Flags: &types.GoAwayFlags{},
					Payload: &types.GoAwayPayload{
						LastStreamID: c.lastReceivedFrame.ID,
						ErrorCode: errors.ProtocolError,
					},
					Length: 8,
				}
				c.outChanFrame <- goaway
				continue
			}
			if newFrame.Flags.(*types.DataFlags).EndStream {
				stream.state = HalfClosedRemote
				c.inChan <- stream
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
			newStream := &Stream{
				id:               newFrame.ID,
				state: 			  streamState,
				lastFrame:        &newFrame,
				headers:          newFrame.Payload.(*types.HeadersPayload).Fragment,
				streamDependency: newFrame.Payload.(*types.HeadersPayload).StreamDependency,
				priorityWeight:   newFrame.Payload.(*types.HeadersPayload).PriorityWeight,
			}
			c.streams[newFrame.ID] = newStream
			if newStream.state == HalfClosedRemote {
				c.inChan <- newStream
			}
		case frame.PriorityType:
			stream, ok := c.streams[newFrame.ID]
			if newFrame.ID == 0 {	
				// Error, a priority frame should be .... with a stream
			}
			if !ok {
				stream = &Stream{
					id: newFrame.ID,
					state: Idle,
					lastFrame: &newFrame,
					headers: make([]byte, 0),
				}
				c.streams[newFrame.ID] = stream
			}
			stream.priorityWeight = newFrame.Payload.(*types.PriorityPayload).PriorityWeight
			stream.streamDependency = newFrame.Payload.(*types.PriorityPayload).StreamDependency
		case frame.RstStreamType:
			stream, ok := c.streams[newFrame.ID]
			if !ok {
				// Error
			}
			stream.state = Closed
			// TODO HANDLE ERROR CODE SENT IN FRAME
		case frame.SettingsType:
		case frame.PushPromiseType:
		case frame.PingType:
			if newFrame.ID != 0 || newFrame.Length != 8 {
				// ERROR
				pingFrame := &frame.Frame{
					ID: 0,
					Type: frame.PingType,
					Length: 8,
					Flags: &types.PingFlags{
						Ack: true,
					},
					Payload: newFrame.Payload,
				}
				// TODO: Might crash, check if tlsConn is blocking
				c.tlsConn.Write(pingFrame.ToBytes())
			}
			if !newFrame.Flags.(*types.PingFlags).Ack {

			}
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
			if newFrame.Flags.(*types.ContinuationFlags).EndHeaders {
				stream.state = Open
			}
			stream.headers = append(stream.headers, newFrame.Payload.(*types.ContinuationPayload).HeaderFragment...)
		}
		c.lastReceivedFrame = newFrame
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
