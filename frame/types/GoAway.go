package types

import (
	"encoding/binary"
	"io"
)

type GoAwayFlags struct{}

func (g GoAwayFlags) ReadFlags(flags byte) {}

type GoAwayPayload struct {
	LastStreamID uint32
	ErrorCode    uint32
	DebugData    []byte
}

func (g GoAwayPayload) ReadPayload(r io.Reader, length uint32, flags IFlags) {
	idBuffer := make([]byte, 4)
	r.Read(idBuffer)
	g.LastStreamID = binary.BigEndian.Uint32(idBuffer) & 0x8000

	errBuffer := make([]byte, 4)
	r.Read(errBuffer)
	g.ErrorCode = binary.BigEndian.Uint32(errBuffer)

	dataBuffer := make([]byte, length-8)
	r.Read(dataBuffer)

	g.DebugData = dataBuffer
}

type GoAway struct {
	Flags   GoAwayFlags
	Payload GoAwayPayload
}

func CreateGoAway(flags byte, payload io.Reader, payloadLength uint32) *GoAway {
	goAway := &GoAway{}
	goAway.Flags.ReadFlags(flags)
	goAway.Payload.ReadPayload(payload, payloadLength, goAway.Flags)

	return goAway
}
