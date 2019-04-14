package types

import (
	"encoding/binary"
	"io"
)

type PriorityFlags struct{}

func (p PriorityFlags) ReadFlags(flags byte) {}

type PriorityPayload struct {
	StreamExclusive  bool
	StreamDependency uint32
	PriorityWeight   byte
}

func (p PriorityPayload) ReadPayload(r io.Reader, length uint32, flags IFlags) {
	streamDependencyBuffer := make([]byte, 4)
	r.Read(streamDependencyBuffer)

	p.StreamExclusive = (streamDependencyBuffer[0] & 0x80) != 0x00
	p.StreamDependency = binary.BigEndian.Uint32(streamDependencyBuffer) & 0x8000

	weightBuffer := make([]byte, 1)
	r.Read(weightBuffer)

	p.PriorityWeight = weightBuffer[0]
}

type Priority struct {
	Flags   PriorityFlags
	Payload PriorityPayload
}

func CreatePriority(flags byte, payload io.Reader, payloadLength uint32) *Priority {
	priority := &Priority{}
	priority.Flags.ReadFlags(flags)
	priority.Payload.ReadPayload(payload, payloadLength, priority.Flags)

	return priority
}
