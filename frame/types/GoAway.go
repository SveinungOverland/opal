package types

import (
	"encoding/binary"
)

type GoAwayFlags struct{}

func (g GoAwayFlags) ReadFlags(flags byte) {}

func (g GoAwayFlags) Byte() (flags byte) {
	return
}

type GoAwayPayload struct {
	LastStreamID uint32
	ErrorCode    uint32
	DebugData    []byte
}

func (g *GoAwayPayload) ReadPayload(payload []byte, length uint32, flags IFlags) {
	g.LastStreamID = binary.BigEndian.Uint32(payload[0:4]) & 0x8000
	g.ErrorCode = binary.BigEndian.Uint32(payload[4:8])
	g.DebugData = payload[8:]
}

func (g GoAwayPayload) Bytes(flags IFlags) []byte {
	buffer := make([]byte, 8+len(g.DebugData))
	binary.BigEndian.PutUint32(buffer[:4], g.LastStreamID)
	binary.BigEndian.PutUint32(buffer[4:8], g.ErrorCode)
	copy(buffer[8:], g.DebugData)
	return buffer
}

type GoAway struct {
	Flags   GoAwayFlags
	Payload GoAwayPayload
}

func CreateGoAway(flags byte, payload []byte, payloadLength uint32) *GoAway {
	goAway := &GoAway{}
	goAway.Flags.ReadFlags(flags)
	goAway.Payload.ReadPayload(payload, payloadLength, goAway.Flags)

	return goAway
}
