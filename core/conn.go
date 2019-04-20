package core

import (
	"crypto/tls"
	"fmt"
	"net"
	"opal/frame/types"
	"opal/hpack"

	"opal/frame"
)

type StreamState uint8

const (
	idle StreamState = iota + 1
	reservedLocal
	reservedRemote
	open
	halfClosedLocal
	halfClosedRemote
	closed
)

type Stream struct {
	id        uint32
	lastFrame *frame.Frame
	state     StreamState
}

type Conn struct {
	server        *Server
	conn          net.Conn
	tlsConn       *tls.Conn
	hpack         *hpack.Context
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
	settingsFrame := frame.ReadFrame(c.tlsConn)
	fmt.Printf("%+v\n", settingsFrame)

	// TODO: Change actual settings based on the frame above
	settingsResponse := &frame.Frame{
		ID:     0,
		Type:   frame.SettingsType,
		Length: 0,
		Flags: types.SettingsFlags{
			Ack: true,
		},
		Payload: &types.SettingsPayload{},
	}

	fmt.Printf("%+v\n", settingsResponse)

	c.tlsConn.Read(prefaceBuffer)
	fmt.Println(string(prefaceBuffer))

	// TODO: Write settingsResponse to client to acknowledge settings frame

	// // Listen for frames
	// for {
	// 	frame.ReadFrame(c.conn)
	// }
}
