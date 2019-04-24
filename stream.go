package opal

import (
	"opal/frame"
	"opal/hpack"
	"opal/http"
	"strings"
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
	id               uint32
	streamDependency uint32
	priorityWeight   byte
	lastFrame        *frame.Frame
	state            StreamState
	headers          []byte
	data             []byte
	endHeaders       bool
	endStream        bool
}

// Build builds and returns a Request based on recieved headers and data frames
func (s *Stream) Build(context *hpack.Context) (*http.Request, error) {
	// Merge and Decode headers
	decoded, err := context.Decode(s.headers) // Header decompression
	if err != nil {
		return nil, err
	}

	// Build request
	req := http.NewRequest()

	// Parse Headers
	for _, hf := range decoded {
		if strings.HasPrefix(hf.Name, ":") {
			req.parsePseudoHeader(hf.Name, hf.Value)
		} else {
			req.Header[hf.Name] = hf.Value
		}
	}

	req.Body = s.data

	return req, nil
}

// ResToFrames converts a response to an array of frames
func (s *Stream) ResToFrames(res *http.Response, context *hpack.Context) ([]frame.Frame, error) {
	return nil, nil
}
