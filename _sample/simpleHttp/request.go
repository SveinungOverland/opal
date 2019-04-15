package simpleHttp

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
)

type Request struct {
	Method     string
	RequestURI string
	RawQuery   string
	Header     map[string]string
	Body       []byte
}

// ---- METHODS -----

func (r Request) Json(target interface{}) error {
	return json.Unmarshal(r.Body, target)
}

// ---- INITIALIZERS -----
// Reads the Request and stores the headers
func readRequest(r io.Reader) (Request, error) {
	req := Request{Header: make(map[string]string)}
	reader := bufio.NewReader(r)

	// Read Request line
	bytes, err := reader.ReadBytes('\n')
	if err != nil {
		return req, err
	}
	err = req.parseRequestLine(string(bytes))
	if err != nil {
		return req, err
	}

	// Reader headers
	for {
		bytes, err := reader.ReadBytes('\n')
		line := string(bytes)
		if line == "\r\n" || line == "" {
			break
		}
		err = req.parseHeaderLine(line)
		if err != nil {
			return req, err
		}
	}

	// Read body
	if lengthValue, hasLength := req.Header["content-length"]; hasLength {
		length, err := strconv.Atoi(lengthValue)
		if err != nil {
			return req, err
		}
		err = req.readBody(reader, length)
		if err != nil {
			return req, err
		}
	}
	return req, nil
}

// ----- HELPERS -------

// ParseRequestLine parses the first line of the Request
// and extracts the method and the route
func (r *Request) parseRequestLine(line string) error {
	RequestLineInfo := strings.Fields(line)
	if len(RequestLineInfo) > 2 {
		r.Method = RequestLineInfo[0]

		route := strings.SplitN(RequestLineInfo[1], "?", 2)
		r.RequestURI = route[0]

		// Parse raw query
		if len(route) > 1 {
			r.RawQuery = route[1]
		}
		return nil
	}
	return errors.New("Invalid HTTP Request")
}

// ParseHeaderLine parses a header line
func (r *Request) parseHeaderLine(line string) error {
	headerKeyPair := strings.SplitN(line, ": ", 2)
	if len(headerKeyPair) != 2 {
		return errors.New("Header-format was incorrect and could not be parsed")
	}

	if r.Header == nil {
		r.Header = make(map[string]string)
	}

	r.Header[strings.ToLower(strings.Trim(headerKeyPair[0], " "))] = strings.Trim(headerKeyPair[1], "\r\n")
	return nil
}

// Reads the body
func (r *Request) readBody(io *bufio.Reader, length int) error {
	// Reads body
	buffer := make([]byte, length)
	_, err := io.Read(buffer)
	if err != nil {
		return err
	}
	r.Body = buffer
	return nil
}
