package http

type Response struct {
	Status uint16
	Body   []byte
	Header map[string]string
}

func (r *Response) NotFound() {
	r.Status = 404
}

func NewResponse() *Response {
	res := &Response{
		Status: 200,
		Body:   make([]byte, 0),
		Header: make(map[string]string),
	}
	res.Header["Content-Type"] = "text/plain"
	return res
}

func new404Response() *Response {
	res := NewResponse()
	res.Status = 404
	return res
}
