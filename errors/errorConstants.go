package errors

const (
	// NoError : The associated condition is not a result of an error. For example, a GOAWAY might include this code to indicate graceful shutdown of a connection.
	NoError uint32 = iota
	// ProtocolError : The endpoint detected an unspecific protocol error. This error is for use when a more specific code is not available.
	ProtocolError
	// InternalError : The enpoint encountered an unexpected internal error.
	InternalError
	// FlowControlError : The endpoint detected that its peer violated the flow-control protocol
	FlowControlError
	// SettingsTimeout : The endpoint sent a SETTINGS frame but did not receive a response in a timely manner. See "Settings Synchronization"
	SettingsTimeout
	// StreamClosed : The endpoint received a frame after a stream was half-closed
	StreamClosed
	// FrameSizeError : The endpoint received a frame with an invalid size
	FrameSizeError
	// RefusedStream : The endpoint refused the stream prior to performing any application processing
	RefusedStream
	// Cancel : Used by the endpoint to indicate that the stream is no longer needed
	Cancel
	// CompressionError : The enpoint is unable to maintain the header compression context for the connection
	CompressionError
	// ConnectError : The connection established in response to a CONNECT request was reset or abnormally closed
	ConnectError
	// EnhanceYourCalm : The endpoint detected that its peer is exhibiting a behavior that might be generating excessive load
	EnhanceYourCalm
	// InadequateSecurity : The underlying transport has properties that do not meet minimum security requirements
	InadequateSecurity
	// HTTP11Required : The endpoint requires that HTTP/1.1 be used instead of HTTP/2 (pfffh)
	HTTP11Required
)

// Any other error should be treated as InternalError