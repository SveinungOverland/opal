package opal

import (
	"fmt"
	"opal/http"
	"opal/router"
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

func serveStreamHandler(conn *Conn) {
	// A channel to handle finished requests (responses)
	reqDoneChan := make(chan responseWrapper, 10)
	defer close(reqDoneChan)

	for {
		select {
		// Check for and handle incoming stream
		case s := <- conn.inChan:
			req, err := createRequest(conn, s) // Header decompression, creating request and response
			if err != nil {
				fmt.Println(err)
				continue
			}
			go serveRequest(conn, reqDoneChan, req, s.id) // Serve and build response

		// Check for and handle incoming responses
		case resWrp := <- reqDoneChan:
			stream, found := conn.streams[resWrp.streamID]
			if !found {
				fmt.Printf("Stream %d is no longer available in stream-map. Throwing away response.\n", resWrp.streamID)
			}
			sendResponse(conn, stream, resWrp.res)
		}
	}
}

// CreateRequest takes a stream and builds a request out of it. Internally it performs a header decompression.
func createRequest(conn *Conn, s *Stream) (*http.Request, error) {
	// Build request
	req, err := s.toRequest(conn.hpack) // Header decompression
	if err != nil {
		return nil, err
	}

	return req, nil
}

// ServeRequest handles an incoming request. Runs all endpoint-methods 
func serveRequest(conn *Conn, reqDoneChan chan responseWrapper, req *http.Request, streamID uint32) {
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

	reqDoneChan <- responseWrapper{res, streamID}
}


func sendResponse(conn *Conn, s *Stream, res *http.Response) error {
	
	// TODO: Implement push promise requests
	_ := res.PushRequests()

	// Decode headers
	decodedHeaders, err := conn.hpack.EncodeMap(res.Header) // Header compression
	if err != nil {
		return err
	}
	s.headers = decodedHeaders
	s.data = res.Body
	
	// Send stream to outChannel
	conn.outChan <- s
	return nil
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