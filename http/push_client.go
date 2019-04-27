package http

type PushClient struct {
	req *Request
	res *Response
}

// NewPusher creates a new PushClient for sending push-requests
func NewPusher(req *Request, res *Response) (*PushClient) {
	return &PushClient {
		req: req,
		res: res,
	}
}

func (pc *PushClient) Push(path string) {
	// Build new request
	req := NewRequest()
	req.Method = "GET"
	req.URI = path
	req.Scheme = pc.req.Scheme

	// RFC7540 claims that a ":authority" pseudo-header must be sent where the server is authoriative
	req.Authority = pc.req.Authority 
	
	if (pc.req.OnPush != nil) {
		pc.req.OnPush(req)
	}
}



