package http

import (
	"opal/hpack"
	"strings"
)

// Request contains and manages all request-relevant data.
type Request struct {
	Method    string
	URI       string
	Authority string // Host
	Scheme    string
	RawQuery  string
	Params    map[string]string
	Header    map[string]string
	Body      []byte

	next   bool   // Bool for deciding if next request can be handled
	Reject func() // Changes the next-value to true
}

// ----- PRIVATE METHODS ------

// Parses HTTP2 Psuedo-Request-Header fields that starts with ":".
func (r *Request) parsePseudoHeader(headerName string, value string) {
	switch headerName {
	case ":authority":
		r.Authority = value
	case ":method":
		r.Method = value
	case ":path":
		uriValues := strings.SplitN(value, "?", 2)
		r.URI = uriValues[0]
		if len(uriValues) > 1 {
			r.RawQuery = uriValues[1]
		}
	case ":scheme":
		r.Scheme = value
	}
}

// NewRequest builds a new request with initialized fields
func NewRequest() *Request {
	req := &Request{
		next:   true,
		Body:   make([]byte, 0),
		Header: make(map[string]string),
		Params: make(map[string]string),
	}
	req.Reject = func() { req.next = false }
	return req
}

// BuildRequest parses and returns a new Request based on given headers and payload
func BuildRequest(hfs []*hpack.HeaderField, body []byte) *Request {
	req := NewRequest()

	// Parse Headers
	for _, hf := range hfs {
		if strings.HasPrefix(hf.Name, ":") {
			req.parsePseudoHeader(hf.Name, hf.Value)
		} else {
			req.Header[hf.Name] = hf.Value
		}
	}

	req.Body = body
	return req
}
