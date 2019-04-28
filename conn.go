package opal

import (
	"crypto/tls"
	"fmt"
	"net"
	"opal/hpack"

	"opal/frame"
	"opal/frame/types"
	"opal/errors"

	error "errors"
	"context"
	"sync"
	"strings"
)

const initialHeaderTableSize = uint32(4096)
var streamMapMutex = sync.Mutex{}

// Conn represents a HTTP-connection
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

// SetStream sets the stream
func (c *Conn) SetStream(s *Stream) {
	streamMapMutex.Lock()
	defer streamMapMutex.Unlock()
	c.streams[s.id] = s
}

// GetStream gets a stream
func (c *Conn) GetStream(id uint32) (*Stream, bool) {
	streamMapMutex.Lock()
	defer streamMapMutex.Unlock()
	s, ok := c.streams[id]
	return s, ok
}

func (c *Conn) serve() {
	// start := time.Now() // Request timer
	defer c.cancel()
	defer close(c.inChan)
	defer close(c.outChan)
	defer close(c.outChanFrame)


	// Helper funcs
	NewConnErr := func(connErr uint32) *frame.Frame {
		return &frame.Frame{
			ID: 0,
			Type: frame.GoAwayType,
			Flags: &types.GoAwayFlags{},
			Payload: &types.GoAwayPayload{
				LastStreamID: c.lastReceivedFrame.ID,
				ErrorCode: connErr,
			},
			Length: 8,
		}
	}

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
	// TODO: Check prefaceBuffer ^^
	if !strings.HasPrefix(string(prefaceBuffer), "PRI") {
		fmt.Println("Invalid HTTP/2 preface-buffer: " + string(prefaceBuffer))
		return
	}

	settingsFrame, err := frame.ReadFrame(c.tlsConn)
	if err != nil {
		c.server.NonBlockingErrorChanSend(err)
	}
	
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

	settingsResponse := &frame.Frame{
		ID:     0,
		Type:   frame.SettingsType,
		Length: 0,
		Flags: &types.SettingsFlags{
			Ack: true,
		},
	}

	go serveStreamHandler(c) // Starting go-routine that is responsible for handling requests when streams are done
	go WriteStream(c) // Starting go-routine that is responsible for handling handled requests that should be written back to client
	
	c.outChanFrame <- settingsResponse
	
	// Connection initiated and ready to receive header frames
	// errors.EnhanceYourCalm
	loop: for {
		select {
		case <-c.ctx.Done():
			break loop
		default:
		}

		newFrame, err := frame.ReadFrame(c.tlsConn)
		if err != nil {
			c.server.NonBlockingErrorChanSend(err)
			break loop
		}

		switch newFrame.Type {
		case frame.DataType:
			// Data should always be associated with a stream
			stream, ok := c.GetStream(newFrame.ID)
			if !ok {
				continue loop
			}
			if !(stream.state == Open || stream.state == HalfClosedLocal) {
				// Stream is not in a state where it can receive data frames
				c.outChanFrame <- frame.NewErrorFrame(stream.id, errors.StreamClosed)
				continue loop
			}
			if stream.id == 0 {
				c.outChanFrame <- NewConnErr(errors.ProtocolError)
				continue loop
			}
			if stream.data == nil {
				stream.data = newFrame.Payload.(*types.DataPayload).Data
			} else {
				stream.data = append(stream.data, newFrame.Payload.(*types.DataPayload).Data...)
			}
			if newFrame.Flags.(*types.DataFlags).EndStream {
				stream.state = HalfClosedRemote
				c.inChan <- stream
			}
		case frame.HeadersType:
			// New stream
			if newFrame.ID == 0 {
				// Error, a header should always be associated with a stream
				c.outChanFrame <- NewConnErr(errors.ProtocolError)
				continue loop
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
			c.SetStream(newStream)
			if newStream.state == HalfClosedRemote {
				c.inChan <- newStream
			}
		case frame.PriorityType:
			stream, ok := c.GetStream(newFrame.ID)
			if newFrame.ID == 0 {	
				c.outChanFrame <- NewConnErr(errors.ProtocolError)
				continue loop
			}
			if newFrame.Length != 5 {
				c.outChanFrame <- frame.NewErrorFrame(stream.id, errors.FrameSizeError)
				continue loop
			}
			if !ok {
				stream = &Stream{
					id: newFrame.ID,
					state: Idle,
					lastFrame: &newFrame,
					headers: make([]byte, 0),
				}
			}
			stream.priorityWeight = newFrame.Payload.(*types.PriorityPayload).PriorityWeight
			stream.streamDependency = newFrame.Payload.(*types.PriorityPayload).StreamDependency
		case frame.RstStreamType:
			stream, ok := c.GetStream(newFrame.ID)
			if newFrame.ID == 0 {
				c.outChanFrame <- NewConnErr(errors.ProtocolError)
				continue loop
			}
			if newFrame.Length != 4 {
				c.outChanFrame <- frame.NewErrorFrame(stream.id, errors.FrameSizeError)
				continue loop
			}
			if !ok {
				c.outChanFrame <- NewConnErr(errors.ProtocolError)
				continue loop
			}
			stream.state = Closed
			// TODO HANDLE ERROR CODE SENT IN FRAME
		case frame.SettingsType:
			if newFrame.ID != 0 {
				c.outChanFrame <- NewConnErr(errors.ProtocolError)
				continue loop
			}
			if newFrame.Flags.(*types.SettingsFlags).Ack {
				if newFrame.Length != 0 {
					c.outChanFrame <- NewConnErr(errors.FrameSizeError)
				}
				continue loop
			}
			if newFrame.Length % 6 != 0 {
				c.outChanFrame <- NewConnErr(errors.FrameSizeError)
				continue loop
			}
			if settingsFrame.Length > 0 {
				for key, value := range settingsFrame.Payload.(*types.SettingsPayload).IDValuePair {
					if key >= 0x1 && key <= 0x6 {
						// Any other key is out of range and is ignored
						c.settings[key] = value
					}
				}
			}
			settingsResponse := &frame.Frame{
				ID:     0,
				Type:   frame.SettingsType,
				Length: 0,
				Flags: &types.SettingsFlags{
					Ack: true,
				},
			}
			c.outChanFrame <- settingsResponse
		case frame.PushPromiseType:
			// Server does not handle PushPromises
		case frame.PingType:
			if newFrame.ID != 0 {
				c.outChanFrame <- NewConnErr(errors.ProtocolError)
				continue loop
			}
			if newFrame.Length != 8 {
				c.outChanFrame <- NewConnErr(errors.FrameSizeError)
				continue loop
			}
			if !newFrame.Flags.(*types.PingFlags).Ack {
				pingResponse := &frame.Frame{
					ID: 0,
					Type: frame.PingType,
					Flags: &types.PingFlags{
						Ack: true,
					},
					Payload: newFrame.Payload,
					Length: 8,
				}
				c.outChanFrame <- pingResponse
			}
		case frame.GoAwayType:
			if newFrame.ID != 0 {
				c.outChanFrame <- NewConnErr(errors.ProtocolError)
				continue loop
			}
			if newFrame.Payload.(*types.GoAwayPayload).ErrorCode != errors.NoError {
				c.server.NonBlockingErrorChanSend(error.New(fmt.Sprint(newFrame.Payload.(*types.GoAwayPayload).ErrorCode)))
				break loop
			} else {
				c.cancel()
				continue loop
			}
		case frame.WindowUpdateType:
			// Update the window size
			if newFrame.Payload.(*types.WindowUpdatePayload).WindowSizeIncrement == 0 {
				c.outChanFrame <- NewConnErr(errors.ProtocolError)
				continue loop
			}
			if newFrame.Length != 4 {
				c.outChanFrame <- NewConnErr(errors.FrameSizeError)
				continue loop
			}
			if newFrame.ID == 0 {
				c.windowSize += newFrame.Payload.(*types.WindowUpdatePayload).WindowSizeIncrement
			}
		case frame.ContinuationType:
			// Append headerfragment
			stream, ok := c.GetStream(newFrame.ID)
			if !ok || newFrame.ID == 0 {
				// Error continuation should always only follow a header
				c.outChanFrame <- NewConnErr(errors.ProtocolError)
			}
			if newFrame.Flags.(*types.ContinuationFlags).EndHeaders {
				stream.state = Open
			}
			stream.headers = append(stream.headers, newFrame.Payload.(*types.ContinuationPayload).HeaderFragment...)
		}
		c.lastReceivedFrame = &newFrame
	}
}


