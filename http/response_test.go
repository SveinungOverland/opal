package http

import (
	"encoding/json"
	"github.com/go-test/deep"
	"testing"
)

func TestResponseStatus(t *testing.T) {
	res := NewResponse(nil)
	checkResStatus(t, res, 200)
	res.BadRequest()
	checkResStatus(t, res, 400)
	res.Unauthorized()
	checkResStatus(t, res, 401)
	res.Forbidden()
	checkResStatus(t, res, 403)
	res.NotFound()
	checkResStatus(t, res, 404)
	res.Ok()
	checkResStatus(t, res, 200)
	res.Created()
	checkResStatus(t, res, 201)
}

func TestResponseJSON(t *testing.T) {
	res := NewResponse(nil)

	// Initialize data
	var result struct {
		A int
		B string
		C bool
	}
	result.A = 200
	result.B = "TEST"
	result.C = false
	expected, _ := json.Marshal(&result)

	// Parse JSON
	res.JSON(200, &result)

	// Validate
	if diff := deep.Equal(expected, res.Body); diff != nil {
		t.Error(diff)
	}
}

// -------- HELPERS ----------
func checkResStatus(t *testing.T, res *Response, expected uint16) {
	if res.Status != expected {
		t.Errorf("Response has incorrect status code. Expected %d, got %d", expected, res.Status)
	}
}
