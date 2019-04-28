package http

import (
	"testing"
)

func TestRequestQuery(t *testing.T) {
	req := NewRequest()
	req.RawQuery = "?name=anders&lastName=Iversen"

	testQuery(t, req, "name", "anders")
	testQuery(t, req, "lastName", "Iversen")

	req.RawQuery = "?tokentoken=asdfasdf%20asdfasdf%20&token=mytoken"
	testQuery(t, req, "token", "mytoken")
	testQuery(t, req, "tokentoken", "asdfasdf%20asdfasdf%20")
}

// -------- HELPERS -------
func testQuery(t *testing.T, req *Request, name, expected string) {
	actual := req.Query(name)
	if actual != expected {
		t.Errorf("Incorrect query value. Expected %s, got %s", expected, actual)
	}
}
