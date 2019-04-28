package opal

import (
	"crypto/tls"
	"fmt"
	"net"
	"opal/router"
	"opal/frame"

	"context"
)

// Server represents a HTTP-server
type Server struct {
	cert          tls.Certificate
	isTLS         bool
	connErrorChan *chan error
	rootRoute     *router.Route
}

// NewTLSServer creates a new http2-server with a TLS configuration
func NewTLSServer(certPath, privateKeyPath string, errorChannel *chan error) (*Server, error) {
	cert, err := tls.LoadX509KeyPair(certPath, privateKeyPath)
	if err != nil {
		return nil, err
	}

	return &Server{
		cert:          cert,
		isTLS:         true,
		connErrorChan: errorChannel,
		rootRoute:     router.NewRoot(),
	}, nil
}

// Listen establishes a TCP-listener on a given port
func (s *Server) Listen(port int16) error {
	fmt.Println("Starting http2 server on port", port)

	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.nonBlockingErrorChanSend(err)
			continue
		}

		c := s.createConn(conn)
		fmt.Printf("New connection established: %s\n", conn.RemoteAddr().String())
		go c.serve()
	}
}

// Register registeres a router to the server
func (s *Server) Register(r *router.Router) {
	s.rootRoute.AppendRouter(r)
}

func (s *Server) createConn(conn net.Conn) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Conn{
		ctx: ctx,
		cancel: cancel,
		server: s,
		conn:   conn,
		isTLS:  false,
		streams: make(map[uint32]*Stream),
		inChan: make(chan *Stream, 10),
		outChan: make(chan *Stream, 1),
		outChanFrame: make(chan *frame.Frame, 1),
		settings: map[uint16]uint32{
			// !ok value should be treated as no-limit
			1: 4096, // Header Table Size
			2: 1, // Enable Push
			//3: no-limit,  // Max Concurrent Streams
			4: 65535, // Initial Window Size
			5: 16384, // Max Frame Size
			//6: no-limit, // Max Header List Size
		},
	}

	if s.isTLS {
		config := &tls.Config{
			Certificates: []tls.Certificate{s.cert},
			ServerName:   "localhost", //Todo: change this
			NextProtos:   []string{"h2"},
		}
		c.tlsConn = tls.Server(conn, config)
		c.isTLS = true
	}

	return c
}

func (s *Server) nonBlockingErrorChanSend(err error) {
	if s.connErrorChan != nil {
		select {
		case *s.connErrorChan <- err:
		default:
			fmt.Println("Error occured but error channel could not receive it, buffer might be full")
		}
	} else {
		fmt.Println("Error occured but error channel does not exist")
	}
}
