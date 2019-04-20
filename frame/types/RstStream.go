package types

import (
	"encoding/binary"
)

type RstStreamFlags struct{}

func (r RstStreamFlags) ReadFlags(flags byte) {}

func (r RstStreamFlags) Byte() (flags byte) {
	return
}

type RstStreamPayload struct {
	ErrorCode uint32 // Should this be uint or int, rfc doc doens't specify uint...
}

func (rst *RstStreamPayload) ReadPayload(payload []byte, length uint32, flags IFlags) {
	rst.ErrorCode = binary.BigEndian.Uint32(payload[:4])
}

func (rst RstStreamPayload) Bytes(flags IFlags) []byte {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, rst.ErrorCode)
	return buffer
}

type RstStream struct {
	Flags   RstStreamFlags
	Payload RstStreamPayload
}

func CreateRstStream(flags byte, payload []byte, payloadLength uint32) *RstStream {
	rstStream := &RstStream{}
	rstStream.Flags.ReadFlags(flags)
	rstStream.Payload.ReadPayload(payload, payloadLength, rstStream.Flags)

	return rstStream
}
