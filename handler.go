package opal

import (
	"fmt"
	"github.com/SveinungOverland/opal/constants"
	"github.com/SveinungOverland/opal/frame"
	"github.com/SveinungOverland/opal/frame/types"
	"github.com/SveinungOverland/opal/hpack"
	"github.com/SveinungOverland/opal/http"
	"github.com/SveinungOverland/opal/router"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
)

// The purpose of this file is to handle streams.
// By this it means building requests, responses, and push-promises,
// and then putting it to the decoder's channel
// This file will also handle all the header compression and decompression (HPACK)

type responseWrapper struct {
	req *http.Request
	res *http.Response
	s   *Stream
}

// ServeStreamHandler handles incoming streams, build and sends responses, and handles push reqests
func serveStreamHandler(conn *Conn) {
	// A channel to handle finished requests (responses)
	reqDoneChan := make(chan responseWrapper, 10)
	defer close(reqDoneChan)

	// Check if server push is enabled
	serverPushEnabled := conn.settings[2] != 0

	for {
		select {
		// Check if connection is done, if so, return
		case <-conn.ctx.Done():
			return

		// Check for and handle incoming stream
		case s := <-conn.inChan:
			if s == nil {
				return
			}
			req, err := createRequest(conn, s) // Header decompression, creating request
			if err != nil {
				fmt.Println(err)
				conn.outChanFrame <- frame.NewErrorFrame(s.id, constants.CompressionError)
				continue
			}

			go handleRequest(conn, reqDoneChan, req, s) // Serve and build response

		// Check for and handle incoming responses
		case resWrp := <-reqDoneChan:
			// Initialize server push requests
			pushRequests := resWrp.res.PushRequests()
			var pushResponses []*responseWrapper
			if serverPushEnabled {
				pushResponses = sendPushRequest(conn, pushRequests, resWrp.s)
			}

			// Send original response
			sendResponse(conn, resWrp.s, resWrp.res)

			// Send push responses
			if serverPushEnabled {
				for _, pshResWrp := range pushResponses {
					sendResponse(conn, pshResWrp.s, pshResWrp.res)
				}
			}
		}
	}
}

// ------------ REQUEST FUNCTIONS ---------------

// CreateRequest takes a stream and builds a request out of it. Internally it performs a header decompression.
func createRequest(conn *Conn, s *Stream) (*http.Request, error) {
	// Build request
	req, err := s.toRequest(conn.hpack) // Header decompression
	if err != nil {
		return nil, err
	}

	return req, nil
}

// HandleRequest builds a response based on given request and sends it to provided out-channel
func handleRequest(conn *Conn, reqDoneChan chan responseWrapper, req *http.Request, s *Stream) {
	res := serveRequest(conn, req)
	reqDoneChan <- responseWrapper{req, res, s}
	go printResponse(req, res)
}

// ServeRequest handles an incoming request. Runs all endpoint-methods
func serveRequest(conn *Conn, req *http.Request) *http.Response {
	// Build response
	res := http.NewResponse(req)

	// Find route and build response
	match, route, params, fh := conn.server.rootRoute.Search(req.URI)
	if match {
		// Found route, run handlers and build response
		req.Params = params
		handlers := route.GetHandlers(req.Method)
		if len(handlers) > 0 {
			middlewares := conn.server.middlewares
			handleRoute(append(middlewares, handlers...), req, res)
		} else {
			res.NotFound() // No handlers found - 404
		}
	} else if fh != nil { // Handle static file response
		handleFile(res, fh)
	} else {
		res.NotFound() // Neither route or file found
	}

	// Set Content-Length if body is provided
	contentLength := len(res.Body)
	if contentLength > 0 {
		res.Header["content-length"] = strconv.Itoa(contentLength)
	}
	return res
}

// ------------ RESPONSE FUNCTIONS ---------------

// SendResponse encodes a response and sends it to a connection's out-channel
func sendResponse(conn *Conn, s *Stream, res *http.Response) {
	// Initialize headers to send
	hfs := make([]*hpack.HeaderField, 0)
	hfs = append(hfs, &hpack.HeaderField{Name: ":status", Value: strconv.Itoa(int(res.Status))})

	for k, v := range res.Header {
		hfs = append(hfs, &hpack.HeaderField{Name: strings.ToLower(k), Value: v})
	}

	// Encode headers
	encodedHeaders := conn.hpack.Encode(hfs) // Header compression

	// Store headers and data in stream
	s.headers = encodedHeaders
	s.data = res.Body
	s.state = HalfClosedRemote

	// Send stream to outChannel
	conn.outChan <- s
}

// ------------ PUSH RESPONSE FUNCTIONS -------------

// SendPushRequest serves a list of requests, builds corresponding responses, and sends new push_promise frames.
// Returns an array containing the responses and created streams.
func sendPushRequest(conn *Conn, reqs []*http.Request, s *Stream) []*responseWrapper {
	pushResponses := make([]*responseWrapper, 0)

	// For all push requests
	for _, pshReq := range reqs {
		// Serve request
		res := serveRequest(conn, pshReq)

		// Build PUSH_PROMISE frame
		pushPromiseFrame := newPushPromise(conn, pshReq, s)

		// Send PUSH_PROMISE frame
		conn.outChanFrame <- pushPromiseFrame

		// Create new stream for request
		stream := &Stream{
			id:    pushPromiseFrame.Payload.(types.PushPromisePayload).StreamID,
			state: ReservedLocal,
		}
		conn.SetStream(stream) // Register stream at conn

		// Append stream and response
		pushResponses = append(pushResponses, &responseWrapper{nil, res, stream})
	}

	return pushResponses
}

// ---- HELPERS -----

// HandleRoute runs endpoint-handlers and builds a response
func handleRoute(handlers []router.HandleFunc, req *http.Request, res *http.Response) {
	for _, handler := range handlers {
		if req.IsFinished() {
			break
		}
		handler(req, res)
	}
}

// HandleFile reads a file and builds a relevant response
func handleFile(res *http.Response, fh *router.FileHandler) {
	file, err := fh.ReadFile()

	// Check if file was found
	if err != nil {
		res.NotFound()
		res.Body = []byte(err.Error())
	} else {
		res.Body = file // File found, return file
		res.Header["content-type"] = fh.MimeType
	}
}

// PushReqHandler creates a new function handler for handling new server push requests
func pushReqHandler(conn *Conn, pushReqChan chan responseWrapper, s *Stream) func(req *http.Request) {
	return func(r *http.Request) {
		res := serveRequest(conn, r)
		pushReqChan <- responseWrapper{r, res, s}
	}
}

// NewPushPromise creates a new PushPromiseFrame based on a given request
func newPushPromise(conn *Conn, req *http.Request, s *Stream) *frame.Frame {

	// Initialize request headers
	hfs := initReqHFs(req)

	// Encode headers
	encodedHeaders := conn.hpack.Encode(hfs) // Header compression
	payloadLength := uint32(len(encodedHeaders))

	// Choose next stream identifier
	// RFC7540 - Section 5.1.1 states that new stream ids from the server must be even
	conn.prevStreamID = conn.prevStreamID + 2 // prevStreamID starts at zero, so it is always even

	pushFrame := &frame.Frame{
		ID:   s.id,
		Type: frame.PushPromiseType,
		Flags: types.PushPromiseFlags{
			EndHeaders: true,
			Padded:     false,
		},
		Payload: types.PushPromisePayload{
			StreamID:  conn.prevStreamID,
			Fragment:  encodedHeaders,
			PadLength: 0,
		},
		Length: payloadLength + 4,
	}

	return pushFrame
}

// InitReqHFs converts a request into a list of headerfields. Pseudo-Headerfields will come first.
func initReqHFs(req *http.Request) []*hpack.HeaderField {
	hfs := make([]*hpack.HeaderField, 0)
	hfs = append(hfs, &hpack.HeaderField{Name: ":method", Value: req.Method})
	hfs = append(hfs, &hpack.HeaderField{Name: ":path", Value: req.URI})
	hfs = append(hfs, &hpack.HeaderField{Name: ":authority", Value: req.Authority})

	if req.Scheme != "" {
		hfs = append(hfs, &hpack.HeaderField{Name: ":scheme", Value: req.Scheme})
	}

	for k, v := range req.Header {
		hfs = append(hfs, &hpack.HeaderField{Name: strings.ToLower(k), Value: v})
	}

	return hfs
}

// PrintResponse prints the request and response status code
var mutex = &sync.Mutex{}

func printResponse(req *http.Request, res *http.Response) {
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
	fmt.Fprintf(color.Output, "HTTP/2 %s %s %s\n", req.Method, req.URI, statusColor(strconv.Itoa(int(res.Status))))
}
