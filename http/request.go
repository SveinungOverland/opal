package http

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

	finished   bool   // Bool for deciding if next request can be handled
	Finish func() // Changes the next-value to true
}

// ----- PRIVATE METHODS ------

// IsFinished says if the request is finished or not 
func (r *Request) IsFinished() bool {
	return r.finished
}

// NewRequest builds a new request with initialized fields
func NewRequest() *Request {
	req := &Request{
		finished:   false,
		Body:   make([]byte, 0),
		Header: make(map[string]string),
		Params: make(map[string]string),
	}
	req.Finish = func() { req.finished = true }
	return req
}
