package opal

import (
	"opal/http"
	"opal/router"
)

// The purpose of this handler is to handle streams.
// By this it means building requests, responses, and push-promises,
// and then putting it to the decoder's channel

func handleStream(conn *Conn, s *Stream) {
	// Build request
	req, err := s.toRequest(conn.hpack)
	if err != nil {
		return 
	}

	// Build response
	res := http.NewResponse()

	// Find route and build response
	match, route, params, fh := conn.server.rootRoute.Search(req.URI)
	if match {
		req.Params = params
		handlers := route.GetHandlers(req.Method)
		if len(handlers) > 0 {
			handleRoute(handlers, req, res)
		} else {
			res.NotFound()
		}
	} else if fh != nil {
		file, err := fh.ReadFile()
		if err != nil {

		} else {
			res.Body = file
		}
	}

	// Build streams
	// Decode headers
	decodedHeaders, err := conn.hpack.EncodeMap(res.Header) // Header compression
	if err != nil {

	}
	s.headers = decodedHeaders
	s.data = res.Body
	
}

func handleRoute(handlers []router.HandleFunc, req *http.Request, res *http.Response) {
	for _, handler := range handlers {
		if req.IsFinished() {
			break
		}
		handler(req, res)
	}
}