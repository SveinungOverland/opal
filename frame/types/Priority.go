package types

import (
	"encoding/binary"
)

type PriorityFlags struct{}

func (p PriorityFlags) ReadFlags(flags byte) {}

func (p PriorityFlags) Byte() (flags byte) {
	return
}

type PriorityPayload struct {
	StreamExclusive  bool
	StreamDependency uint32
	PriorityWeight   byte
}

func (p *PriorityPayload) ReadPayload(payload []byte, length uint32, flags IFlags) {

	p.StreamExclusive = (payload[0] & 0x80) != 0x00
	p.StreamDependency = binary.BigEndian.Uint32(payload[:4]) & 0x7FFF

	p.PriorityWeight = payload[4]
}

func (p PriorityPayload) Bytes(flags IFlags) []byte {
	buffer := make([]byte, 5)
	binary.BigEndian.PutUint32(buffer[:4], p.StreamDependency)
	if p.StreamExclusive {
		buffer[0] |= 0x8
	}
	buffer[4] = p.PriorityWeight

	return buffer
}

type Priority struct {
	Flags   PriorityFlags
	Payload PriorityPayload
}

func CreatePriority(flags byte, payload []byte, payloadLength uint32) *Priority {
	priority := &Priority{}
	priority.Flags.ReadFlags(flags)
	priority.Payload.ReadPayload(payload, payloadLength, priority.Flags)

	return priority
}
