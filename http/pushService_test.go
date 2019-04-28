package http

import (
	"testing"
)

func TestPush(t *testing.T) {
	// Initialize request and response
	authority := "https://example.com"
	req := NewRequest()
	req.Authority = authority
	res := NewResponse(req)

	reqPaths := [3]string{"/test", "/test/home/", "/test/theme.css"}
	for _, path := range reqPaths {
		res.Push(path)
	}

	// Check if push requests has been built
	pushReqs := res.PushRequests()
	for i, req := range pushReqs {
		if req.Method != "GET" {
			t.Errorf("Invalid method in push request. Expected %s, got %s", "GET", req.Method)
		}
		if req.URI != reqPaths[i] {
			t.Errorf("Invalid path in push request. Expected %s, got %s", reqPaths[i], req.URI)
		}
		if req.Authority != authority {
			t.Errorf("Invalid authority in push request. Expected %s, got %s", authority, req.Authority)
		}
	}

}
