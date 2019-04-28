package http

type pushService struct {
	res          *Response
	pushRequests []*Request
}

// NewPusher creates a new PushClient for sending push-requests
func newPushService(res *Response) *pushService {
	return &pushService{
		res:          res,
		pushRequests: make([]*Request, 0),
	}
}

func (ps *pushService) Push(path string) {
	// Build new request
	req := NewRequest()
	req.Method = "GET"
	req.URI = path
	req.Scheme = ps.res.req.Scheme

	// RFC7540 claims that a ":authority" pseudo-header must be sent where the server is authoriative
	req.Authority = ps.res.req.Authority

	ps.pushRequests = append(ps.pushRequests, req)
}
