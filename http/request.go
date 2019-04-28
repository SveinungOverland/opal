package http

import (
	"encoding/json"
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

	finished bool // Bool for deciding if next request can be handled
}

// JSON parses the request body as JSON into a target interface.
func (r *Request) JSON(target interface{}) {
	json.Unmarshal(r.Body, target)
}

// Query returns the value of a given query parameter
func (r *Request) Query(name string) string {
	name = name + "="

	// Find start index
	startIndex := strings.Index(r.RawQuery, "?"+name)
	if startIndex == -1 {
		startIndex = strings.Index(r.RawQuery, "&"+name)
		if startIndex == -1 {
			return ""
		}
	}
	startIndex = startIndex + len(name) + 1

	// Find end index
	endIndex := strings.Index(r.RawQuery[startIndex:], "&")
	if endIndex == -1 {
		endIndex = len(r.RawQuery)
	} else {
		endIndex += startIndex
	}

	return r.RawQuery[startIndex:(endIndex)]
}

// Finish makes the request finished. Which means next handler in the function won't run.
func (r *Request) Finish() {
	r.finished = true
}

// ----- PRIVATE METHODS ------

// IsFinished says if the request is finished or not
func (r *Request) IsFinished() bool {
	return r.finished
}

// NewRequest builds a new request with initialized fields
func NewRequest() *Request {
	req := &Request{
		finished: false,
		Body:     make([]byte, 0),
		Header:   make(map[string]string),
		Params:   make(map[string]string),
	}
	return req
}
