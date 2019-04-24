package opal


// The purpose of this handler is to handle streams.
// By this it means building requests, responses, and push-promises,
// and then putting it to the decoder's channel

func handleStream(conn *Conn, s *Stream) {
	// asdf
	_, err := s.Build(conn.hpack)
	if err != nil {
		return 
	}
}