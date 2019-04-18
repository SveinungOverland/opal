package types

import "io"

/*
	The CONTINUATION frame is used to continue a sequence of header block fragments
*/
type ContinuationFlags struct {
	EndHeaders bool
}

func (c ContinuationFlags) ReadFlags(flags byte) {
	c.EndHeaders = (flags & 0x04) != 0x00
}

type ContinuationPayload struct {
	HeaderFragment []byte
}

func (c ContinuationPayload) ReadPayload(r io.Reader, length uint32, flags IFlags) {
	headerFragmentBuffer := make([]byte, length)
	r.Read(headerFragmentBuffer)

	c.HeaderFragment = headerFragmentBuffer
}

type Continuation struct {
	Flags   ContinuationFlags
	Payload ContinuationPayload
}

func CreateContinuation(flags byte, payload io.Reader, payloadLength uint32) *Continuation {
	continuation := &Continuation{}
	continuation.Flags.ReadFlags(flags)
	continuation.Payload.ReadPayload(payload, payloadLength, continuation.Flags)

	return continuation
}
