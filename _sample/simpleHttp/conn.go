package simpleHttp

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
)

type Conn struct {
	server  *Server
	conn    net.Conn
	Request Request
	tlsConn *tls.Conn
	isTLS   bool
}

func (c *Conn) GetReader() io.Reader {
	if c.isTLS {
		return c.tlsConn
	} else {
		return c.conn
	}
}

func (c *Conn) GetWriter() io.Writer {
	if c.isTLS {
		return c.tlsConn
	} else {
		return c.conn
	}
}

var mutex *sync.Mutex

func init() {
	mutex = &sync.Mutex{}
}

func (c *Conn) serve() {
	defer c.conn.Close()
	start := time.Now() // Time request

	// Initialize TLS Handshake
	if c.isTLS {
		err := c.tlsConn.Handshake()
		if err != nil {
			fmt.Println(err)
			c.isTLS = false
			c.tlsConn.Close()
			return
		}
	}

	// Build and handle request
	req, err := readRequest(c.GetReader())
	if err != nil {
		fmt.Printf("Invalid request from %s -- %s\n", c.conn.RemoteAddr().String(), err.Error())
		return
	}
	c.Request = req

	// Find matching route
	route := findRouteByRequest(c.server.rootNode, &req)

	// Build response
	var res *Response
	if route != nil {
		defaultRes := buildDefaultResponse()
		res = route.Exec(&req, &defaultRes)
	} else {
		notFoundRes := build404Response()
		res = &notFoundRes
	}

	res.write(c.GetWriter())

	duration := time.Since(start) // Calculate request time
	go printResponse(&req, res, &duration)
}

// Prints the request and response status code
func printResponse(req *Request, res *Response, duration *time.Duration) {
	mutex.Lock()
	defer mutex.Unlock()
	var statusColor func(a ...interface{}) string
	if res.Status < 300 {
		statusColor = color.New(color.FgGreen).SprintFunc()
	} else if res.Status < 400 {
		statusColor = color.New(color.FgYellow).SprintFunc()
	} else {
		statusColor = color.New(color.FgRed).SprintFunc()
	}
	fmt.Fprintf(color.Output, "HTTP/1.1 %s %s %s %s\n", req.Method, req.RequestURI, statusColor(strconv.Itoa(res.Status)), duration.String())
}
