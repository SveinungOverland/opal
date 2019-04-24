package opal

import (
	"opal/frame"
	"opal/frame/types"
	"opal/hpack"
	"opal/http"
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
	id        uint32
	lastFrame *frame.Frame
	state     StreamState
	headers   []byte
	data      []byte
}

// Build builds and returns a Request based on recieved headers and data frames
func (s *Stream) Build(context *hpack.Context) (*http.Request, error) {
	// Merge and Decode headers
	var headerBytes []byte
	for _, headers := range s.headers {
		headerBytes = append(headerBytes, headers.Fragment...)
	}
	decoded, err := context.Decode(headerBytes) // Header decompression
	if err != nil {
		return nil, err
	}

	// Merge data
	var data []byte
	for _, dataPayload := range s.data {
		data = append(data, dataPayload.Data...)
	}

	// Build request
	req := http.BuildRequest(decoded, data)

	return req, nil
}

// ResToFrames converts a response to an array of frames
func (s *Stream) ResToFrames(res *http.Response, context *hpack.Context) ([]frame.Frame, error) {
	return nil, nil
}
