package core

import (
	"crypto/tls"
	"fmt"
	"net"

	"opal/frame"
)

type StreamState uint8

const (
	idle StreamState = 1
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

	// Listen for frames
	for {
		frame.ReadFrame(c.conn)
	}
}
