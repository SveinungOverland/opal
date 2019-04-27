package opal

import (
	"fmt"
	"context"
	"opal/errors"
	"opal/http"
	"opal/router"
	"opal/frame"
	"opal/frame/types"
	"opal/hpack"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
)


// The purpose of this file is to handle streams.
// By this it means building requests, responses, and push-promises,
// and then putting it to the decoder's channel
// This file will also do all the header compression and decompression

type responseWrapper struct {
	req		*http.Request
	res 	 *http.Response
	streamID uint32
}

// ServeStreamHandler handles incoming streams, build and sends responses, and handles push reqests
func serveStreamHandler(conn *Conn) {
	// A channel to handle finished requests (responses)
	reqDoneChan := make(chan responseWrapper, 10)
	pushReqChan := make(chan responseWrapper, 10) // Handels server push requests
	defer close(reqDoneChan)
	defer close(pushReqChan)

	// Check if server push is enabled
	serverPushEnabled := conn.settings[2] != 0

	for {
		select {
		// Check if connection is done, if so, return
		case <- conn.ctx.Done():
			return
		// Check for and handle incoming stream
		case s := <- conn.inChan:
			req, err := createRequest(conn, s) // Header decompression, creating request
			if err != nil {
				fmt.Println(err)
				conn.outChanFrame <- frame.NewErrorFrame(s.id, errors.CompressionError)
				continue
			}
			
			// Handle server push
			if (serverPushEnabled) {
				req.OnPush = pushReqHandler(conn, pushReqChan, s.id)
			}
			
			go handleRequest(conn, reqDoneChan, req, s.id) // Serve and build response

		// Check for and handle incoming server push requests
		case pushReqWrp := <- pushReqChan:
			pushPromise := newPushPromise(conn, pushReqWrp.req, pushReqWrp.streamID) // Create Push Request
			conn.outChanFrame <- pushPromise // Send Push Request
			stream := &Stream{
				id: pushReqWrp.streamID,
				state: ReservedLocal,
			}
			sendResponse(conn, stream, pushReqWrp.res) // Send Push Response

		// Check for and handle incoming responses
		case resWrp := <- reqDoneChan:
			stream, found := conn.streams[resWrp.streamID]
			if !found {
				fmt.Printf("Stream %d is no longer available in stream-map. Throwing away response.\n", resWrp.streamID)
				continue
			}
			sendResponse(conn, stream, resWrp.res)
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
func handleRequest(conn *Conn, reqDoneChan chan responseWrapper, req *http.Request, streamID uint32) {
	res := serveRequest(conn, req)
	reqDoneChan <- responseWrapper{req, res, streamID}
	go printResponse(req, res)
}

// ServeRequest handles an incoming request. Runs all endpoint-methods 
func serveRequest(conn *Conn, req *http.Request) *http.Response{
	// Build response
	res := http.NewResponse()

	// Find route and build response
	match, route, params, fh := conn.server.rootRoute.Search(req.URI)
	if match {
		// Found route, run handlers and build response
		req.Params = params
		handlers := route.GetHandlers(req.Method)
		if len(handlers) > 0 {
			handleRoute(handlers, req, res)
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
	if (contentLength > 0) {
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
	
	s.headers = encodedHeaders
	s.data = res.Body

	// Send stream to outChannel
	conn.outChan <- s
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
		res.Header["cache-control"] = "public"
	}
}

// PushReqHandler creates a new function handler for handling new server push requests
func pushReqHandler(conn *Conn, pushReqChan chan responseWrapper, streamID uint32) func(req *http.Request) {
	return func(r *http.Request) {
		res := serveRequest(conn, r)
		pushReqChan <- responseWrapper{r, res, streamID, }
	}
}

// NewPushPromise creates a new PushPromiseFrame based on a given request
func newPushPromise(conn *Conn, req *http.Request, streamID uint32) *frame.Frame {

	// Initialize request headers
	hfs := initReqHFs(req)

	for _, hf := range hfs {
		fmt.Printf("%s - %s\n", hf.Name, hf.Value)
	}

	// Encode headers
	encodedHeaders := conn.hpack.Encode(hfs) // Header compression
	payloadLength := uint32(len(encodedHeaders))

	// Choose next stream identifier
	// RFC7540 - Section 5.1.1 states that new stream ids from the server must be even
	conn.prevStreamID = conn.prevStreamID + 2 // prevStreamID starts at zero, so it is always even
	fmt.Printf("New stream id: %d\n", conn.prevStreamID)
	//pushPromise := types.CreatePushPromise(flags, encodedHeaders, payloadLength)

	pushFrame := &frame.Frame {
		ID: streamID,
		Type: frame.PushPromiseType,
		Flags: types.PushPromiseFlags {
			EndHeaders: true,
			Padded: false,
		},
		Payload: types.PushPromisePayload {
			StreamID: conn.prevStreamID,
			Fragment: encodedHeaders,
			PadLength: 0,
		},
		Length: payloadLength + 4,
	}

	return pushFrame
}

// InitReqHFs converts a request into a list of headerfields. Psuedo-Headerfields will come first.
func initReqHFs(req *http.Request) []*hpack.HeaderField {
	hfs := make([]*hpack.HeaderField, 0)
	hfs = append(hfs, &hpack.HeaderField{Name: ":method", Value: req.Method})
	hfs = append(hfs, &hpack.HeaderField{Name: ":authority", Value: req.Authority})
	hfs = append(hfs, &hpack.HeaderField{Name: ":path", Value: req.URI})

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