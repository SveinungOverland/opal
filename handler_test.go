package opal

import (
	"testing"
	"opal/router"
	"opal/http"
	"opal/hpack"
	"opal/frame"
	"opal/frame/types"
	"crypto/tls"
	"fmt"
	"context"
	"time"
)

// TestStreamHandlerIntegration tests serveStreamHandler function, which will then be an intergration
// test between all the other logic in handler.go
func TestStreamHandlerIntegration(t *testing.T) {
	testCxt, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second) // Context for timeout check

	// Initialize server and connection
	srv := newTestServer()
	srv.Register(newTestRouter())
	conn := srv.createConn(nil)
	conn.hpack = hpack.NewContext(4096, 4096) // Initialize hpack context

	// Initialize HPACK (streams must include encoded headers)
	hpackCtx := hpack.NewContext(4096, 4096)

	// Initialize incoming streams
	s1Headers := newEncodedTestHeaders(hpackCtx, "/", "GET")
	s2Headers := newEncodedTestHeaders(hpackCtx, "/test", "POST")
	s3Headers := newEncodedTestHeaders(hpackCtx, "/invalid", "PUT") // Should result in 404
	s1 := newTestStream(1, s1Headers, []byte{})
	s2 := newTestStream(3, s2Headers, []byte("TEST"))
	s3 := newTestStream(5, s3Headers, []byte{})

	// Initialize StreamHandler
	go serveStreamHandler(conn)

	// Send streams
	conn.inChan <- s1
	conn.inChan <- s2
	conn.inChan <- s3

	// Read incoming streams
	recStreamCount := 0
	recStreams := [3]bool{false, false, false} // s1, s2, s3

	for recStreamCount < 3 {
		select {
		case <- testCxt.Done(): // Timeout check
			t.Error("Not all streams was recieved!")
			break
		case stream := <- conn.outChan:
			recStreamCount++
			headers, err := hpackCtx.Decode(stream.headers)
			if err != nil {
				t.Error(fmt.Sprintf("Decoding error at streem %d", stream.id))
				continue
			}
			if stream.id == 1 {
				recStreams[0] = true
				validateHeaders(t, stream, headers, []*hpack.HeaderField{
					hf(":status", "400"),
					hf("content-type", "text/plain; charset=utf-8"),
				})
				validateData(t, stream, "")
			}
			if stream.id == 3 {
				recStreams[1] = true
				validateHeaders(t, stream, headers, []*hpack.HeaderField{
					hf(":status", "200"),
					hf("content-type", "application/json"),
					hf("content-length", "24"),
				})
				validateData(t, stream, "{\"Result\":\"TEST_RESULT\"}")
			}
			if stream.id == 5 {
				recStreams[2] = true
				validateHeaders(t, stream, headers, []*hpack.HeaderField{
					hf(":status", "404"),
					hf("content-type", "text/plain; charset=utf-8"),
				})
				validateData(t, stream, "")
			}
		}
	}

	// Check if all streams sent are recieved
	for i, recv := range recStreams {
		if !recv {
			t.Errorf("Stream %d was not recieved!", i)
		}
	}

	conn.cancel()
	cancelFunc()
}

func TestSendResponse(t *testing.T) {
	conn := newTestConn()

	// Create response
	res := http.NewResponse(nil)
	res.Body = []byte("TEST_BODY")
	res.Header["Cache-Control"] = "public" // Header should be converted to lowercase
	res.Unauthorized()

	// Send response
	stream := newTestStream(1, []byte{}, []byte{})
	sendResponse(conn, stream, res)
	outStream := <- conn.outChan

	// Check stream id
	if outStream.id != 1 {
		t.Errorf("Invalid stream id recieved! Expected %d, got %d", 1, outStream.id)
	}

	// Check stream headers
	headers, err := conn.hpack.Decode(outStream.headers)
	if err != nil {
		t.Errorf("Decoding error!")
	}
	validateHeaders(t, outStream, headers, []*hpack.HeaderField{
		hf(":status", "401"),
		hf("content-type", "text/plain; charset=utf-8"),
		hf("cache-control", "public"),
	})

	// Validate data
	validateData(t, stream, "TEST_BODY")
}

func TestNewPushPromise(t *testing.T) {
	conn := newTestConn()
	conn.prevStreamID = 4
	// Create new request
	req := http.NewRequest()
	req.Method = "PUT"
	req.URI = "/test"
	req.Authority = "https://example.com"

	stream := newTestStream(4, []byte{}, []byte{})

	pushPromise := newPushPromise(conn, req, stream)

	// Check frame id (shoud be equal to stream id)
	if pushPromise.ID != stream.id {
		t.Errorf("PushPromise frame has invalid id! Expected %d, got %d", stream.id, pushPromise.ID)
	}
	// Check if promised stream identifier
	payload, ok := pushPromise.Payload.(types.PushPromisePayload);
	if !ok {
		t.Error("PushPromiseFrame does not include a PushPromisePayload!")
	}
	if payload.StreamID != 6 { // conn.prevStreamID + 2
		t.Errorf("Incorrect promised stream identifier! Expected %d, got %d", payload.StreamID, 6)
	}
	// Check type
	if pushPromise.Type != frame.PushPromiseType {
		t.Errorf("Incorrect push promise type! Expected %d, got %d", frame.PushPromiseType, pushPromise.Type)
	}
}

// ---------- HELPERS --------------

func newTestServer() (*Server) {
	return &Server{
		cert:          tls.Certificate{},
		isTLS:         true,
		connErrorChan: nil,
		rootRoute:     router.NewRoot(),
	}
}

func newTestConn() *Conn {
	srv := newTestServer()
	conn := srv.createConn(nil)
	conn.hpack = hpack.NewContext(4096, 4096) // Initialize hpack context
	return conn
}

func newTestRouter() (*router.Router) {
	r := router.NewRouter("/")

	r.Get("/", func(req *http.Request, res *http.Response) {
		res.BadRequest()
	})

	r.Post("/test", func(req *http.Request, res *http.Response) {
		var result struct {
			Result string
		}
		result.Result = string(req.Body) + "_RESULT"
		
		res.JSON(result)
	})

	return r
}

func newEncodedTestHeaders(hpackCtx *hpack.Context, path, method string) []byte {
	// Initialize and encode headers
	hfs := []*hpack.HeaderField{
		&hpack.HeaderField{Name: ":method", Value: method},
		&hpack.HeaderField{Name: ":path", Value: path},
	}
	return hpackCtx.Encode(hfs)
}

func newTestStream(id uint32, header, data []byte) (*Stream) {
	return &Stream {
		id: id,
		state: Open,
		headers: header,
		data: data,
	}
}

func validateHeaders(t *testing.T, s *Stream, actual []*hpack.HeaderField, expectedHeaders []*hpack.HeaderField) {
	// Store actual headers in map so the order does not have any effect
	actualHeaders := map[string]string {}
	for _, hf := range actual {
		actualHeaders[hf.Name] = hf.Value
	}

	// Check if expected headers are in actualHeaders map
	for _, hf := range expectedHeaders {
		value, ok := actualHeaders[hf.Name]
		if !ok {
			t.Errorf("Missing header in stream %d! Expected %s!", s.id, hf.Name)
		}
		if hf.Value != value {
			t.Errorf("Incorrect header value in header %s! Expected %s, got %s", hf.Name, hf.Value, value)
		}
	}
}

func validateData(t *testing.T, s *Stream, expected string) {
	actual := string(s.data)
	if actual != expected {
		t.Errorf("Incorrect data recieved from stream %d. Expected %s, got %s!", s.id, expected, actual)
	}
}

func hf(name string, value string) *hpack.HeaderField {
	return &hpack.HeaderField{Name: name, Value: value}
}