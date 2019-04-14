package types

import (
	"encoding/binary"
	"io"
)

type RstStreamFlags struct{}

func (p RstStreamFlags) ReadFlags(flags byte) {}

type RstStreamPayload struct {
	ErrorCode uint32 // Should this be uint or int, rfc doc doens't specify uint...
}

func (rst RstStreamPayload) ReadPayload(r io.Reader, length uint32, flags IFlags) {
	errorCodeBuffer := make([]byte, 4)
	r.Read(errorCodeBuffer)

	rst.ErrorCode = binary.BigEndian.Uint32(errorCodeBuffer)
}

type RstStream struct {
	Flags   RstStreamFlags
	Payload RstStreamPayload
}

func CreateRstStream(flags byte, payload io.Reader, payloadLength uint32) *RstStream {
	rstStream := &RstStream{}
	rstStream.Flags.ReadFlags(flags)
	rstStream.Payload.ReadPayload(payload, payloadLength, rstStream.Flags)

	return rstStream
}
