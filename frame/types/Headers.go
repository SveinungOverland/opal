package types

import (
	"encoding/binary"
	"io"
)

type HeadersFlags struct {
	EndStream  bool
	EndHeaders bool
	Padded     bool
	Priority   bool
}

func (h HeadersFlags) ReadFlags(flags byte) {
	h.EndStream = (flags & 0x01) != 0x00
	h.EndHeaders = (flags & 0x04) != 0x00
	h.Padded = (flags & 0x08) != 0x00
	h.Priority = (flags & 0x20) != 0x00
}

type HeadersPayload struct {
	StreamExclusive  bool   // Can only be set true if Priority flag is set
	StreamDependency uint32 // Can only be present if Priority flag  is set
	PriorityWeight   byte   // Can only be present if Priority flag is set
	Fragment         []byte
}

func (h HeadersPayload) ReadPayload(r io.Reader, length uint32, flags IFlags) {
	padLength := make([]byte, 1)
	bytesToRead := length
	if flags.(HeadersFlags).Padded {
		r.Read(padLength)
		bytesToRead -= uint32(1 + uint8(padLength[0]))
	}
	if flags.(HeadersFlags).Priority {
		streamDependencyBuffer := make([]byte, 4)
		r.Read(streamDependencyBuffer)

		h.StreamExclusive = (streamDependencyBuffer[0] & 0x80) != 0x00
		h.StreamDependency = binary.BigEndian.Uint32(streamDependencyBuffer) & 0x8000

		weightBuffer := make([]byte, 1)
		r.Read(weightBuffer)

		h.PriorityWeight = weightBuffer[0]
	}

	fragmentBuffer := make([]byte, bytesToRead)
	r.Read(fragmentBuffer)

	h.Fragment = fragmentBuffer
}

type Headers struct {
	Flags   HeadersFlags
	Payload HeadersPayload
}

func CreateHeaders(flags byte, payload io.Reader, payloadLength uint32) *Headers {
	headers := &Headers{}
	headers.Flags.ReadFlags(flags)
	headers.Payload.ReadPayload(payload, payloadLength, headers.Flags)

	return headers
}
