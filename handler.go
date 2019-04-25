package opal

import (
	"fmt"
	"opal/http"
	"opal/router"
	"opal/frame"
	"opal/frame/types"
	"strconv"
)


// The purpose of this file is to handle streams.
// By this it means building requests, responses, and push-promises,
// and then putting it to the decoder's channel
// This file will also do all the header compression and decompression

type responseWrapper struct {
	res 	 *http.Response
	streamID uint32
}

type pushReqWrapper struct {
	req		*http.Request
	res 	*http.Response
}

// ServeStreamHandler handles incoming streams, build and sends responses, and handles push reqests
func serveStreamHandler(conn *Conn) {
	// A channel to handle finished requests (responses)
	reqDoneChan := make(chan responseWrapper, 10)
	pushReqChan := make(chan pushReqWrapper, 10) // Handels server push requests
	defer close(reqDoneChan)
	defer close(pushReqChan)

	// Check if server push is enabled
	serverPushEnabled := conn.settings[2] != 0

	for {
		select {
		// Check for and handle incoming stream
		case s := <- conn.inChan:
			req, err := createRequest(conn, s) // Header decompression, creating request
			if err != nil {
				fmt.Println(err)
				continue
			}
			
			// Handle server push
			if (serverPushEnabled) {
				req.OnPush = pushReqHandler(conn, pushReqChan)
			}
			
			go handleRequest(conn, reqDoneChan, req, s.id) // Serve and build response

		// Check for and handle incoming server push requests
		case pushReqWrp := <- pushReqChan:
			pushPromise := newPushPromise(conn, pushReqWrp.req) // Create Push Request
			conn.outChanFrame <- pushPromise // Send Push Request
			stream := &Stream{
				id: pushPromise.ID,
				state: Idle,
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
	reqDoneChan <- responseWrapper{res, streamID}
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
		res.Header["Content-Length"] = strconv.Itoa(contentLength)
	}

	return res
}

// ------------ RESPONSE FUNCTIONS ---------------

// SendResponse encodes a response and sends it to a connection's out-channel
func sendResponse(conn *Conn, s *Stream, res *http.Response) {
	setResPsudeoHeaders(res)

	// Encode headers
	encodedHeaders := conn.hpack.EncodeMap(res.Header) // Header compression
	
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
		res.Header["Content-Type"] = fh.MimeType
	}
}

// PushReqHandler creates a new function handler for handling new server push requests
func pushReqHandler(conn *Conn, pushReqChan chan pushReqWrapper) func(req *http.Request) {
	return func(r *http.Request) {
		res := serveRequest(conn, r)
		pushReqChan <- pushReqWrapper{r, res}
	}
}

// NewPushPromise creates a new PushPromiseFrame based on a given request
func newPushPromise(conn *Conn, req *http.Request) *frame.Frame {
	// Add request pseudo-headers to Header map
	setReqPsuedoHeaders(req)

	// Encode headers
	encodedHeaders := conn.hpack.EncodeMap(req.Header) // Header compression
	payloadLength := uint32(len(encodedHeaders))

	// Choose next stream identifier
	// RFC7540 - Section 5.1.1 states that new stream ids from the server must be even
	conn.prevStreamID = conn.prevStreamID + 2 // prevStreamID starts at zero, so it is always even

	// Decide frame flags
	flags := byte(0x0&64) // END_HEADERS is set. Headers should not be greater than 2^24 bytes anyway

	pushPromise := types.CreatePushPromise(flags, encodedHeaders, payloadLength)

	pushFrame := &frame.Frame {
		ID: conn.prevStreamID,
		Type: frame.PushPromiseType,
		Flags: pushPromise.Flags,
		Payload: pushPromise.Payload,
		Length: payloadLength,
	}

	return pushFrame
}

// SetReqPsuedoHeaders adds the request's pseudo-headers to the header-map
func setReqPsuedoHeaders(req *http.Request) {
	req.Header[":method"] = req.Method
	req.Header[":authority"] = req.Authority
	req.Header[":path"] = req.URI

	if req.Scheme != "" {
		req.Header[":scheme"] = req.Scheme
	}
}

// SetResPsudeoHeaders adds the response's pseudo-headers to the header-map
func setResPsudeoHeaders(res *http.Response) {
	res.Header[":status"] = strconv.Itoa(int(res.Status))
}