package types

/*
	The CONTINUATION frame is used to continue a sequence of header block fragments
*/
type ContinuationFlags struct {
	EndHeaders bool
}

func (c *ContinuationFlags) ReadFlags(flags byte) {
	c.EndHeaders = (flags & 0x4) != 0x0
}

func (c ContinuationFlags) Byte() (flags byte) {
	if c.EndHeaders {
		flags |= 0x4
	}
	return
}

type ContinuationPayload struct {
	HeaderFragment []byte
}

func (c *ContinuationPayload) ReadPayload(payload []byte, length uint32, flags IFlags) {
	c.HeaderFragment = payload
}

func (c ContinuationPayload) Bytes(flags IFlags) []byte {
	return c.HeaderFragment
}

type Continuation struct {
	Flags   ContinuationFlags
	Payload ContinuationPayload
}

func CreateContinuation(flags byte, payload []byte, payloadLength uint32) *Continuation {
	continuation := &Continuation{}
	continuation.Flags.ReadFlags(flags)
	continuation.Payload.ReadPayload(payload, payloadLength, &continuation.Flags)

	return continuation
}
