package opal

import (
	"testing"
	"opal/router"
	"opal/http"
	"opal/hpack"
	"crypto/tls"
	"fmt"
	"github.com/go-test/deep"
)

// TestStreamHandlerIntegration tests serveStreamHandler function, which will then be an intergration
// test between all the other logic in handler.go
func TestStreamHandlerIntegration(t *testing.T) {
	// Initialize server and connection
	srv := newTestServer()
	srv.Register(newTestRouter())
	conn := srv.createConn(nil)
	conn.hpack = hpack.NewContext(4096, 4096) // Initialize hpack context

	// Initialize HPACK
	hpackCtx := hpack.NewContext(4096, 4096)

	// Initialize incoming streams
	s1Headers := newEncodedHeaders(hpackCtx, "/", "GET")
	s2Headers := newEncodedHeaders(hpackCtx, "/test", "POST")
	s3Headers := newEncodedHeaders(hpackCtx, "/invalid", "PUT") // Should result in 404
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
		stream := <- conn.outChan
		recStreamCount++
		headers, err := hpackCtx.Decode(stream.headers)
		if err != nil {
			t.Error(fmt.Sprintf("Decoding error at streem %d", stream.id))
			continue
		}
		fmt.Println(stream)
		if stream.id == 1 {
			recStreams[0] = true
			//validateHeaders(t, stream, headers, )
			if len(headers) == 0 {
				t.Error("Incorrect headers returned for stream 1!")
			}
			statusHeader := headers[0]
			if statusHeader.Value == "200" {
				t.Error(fmt.Sprintf("Incorrect status returned from stream %d! Expected: %s, actual: %s", stream.id, "200", statusHeader.Value))
			}
		}
	}

	conn.cancel()

	

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

func newEncodedHeaders(hpackCtx *hpack.Context, path, method string) []byte {
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
	if diff := deep.Equal(actual, expectedHeaders); diff != nil {
		t.Errorf("Incorrect headers in stream %d. %s", s.id, diff)
	}
}