package types

import (
	"encoding/binary"
)

type PushPromiseFlags struct {
	EndHeaders bool
	Padded     bool
}

func (p *PushPromiseFlags) ReadFlags(flags byte) {
	p.EndHeaders = (flags & 0x4) != 0x0
	p.Padded = (flags & 0x8) != 0x0
}

func (p PushPromiseFlags) Byte() (flags byte) {
	if p.EndHeaders {
		flags |= 0x4
	}
	if p.Padded {
		flags |= 0x8
	}
	return
}

type PushPromisePayload struct {
	StreamID  uint32
	Fragment  []byte
	PadLength byte
}

func (p *PushPromisePayload) ReadPayload(payload []byte, length uint32, flags IFlags) {
	index := 0
	if flags.(*PushPromiseFlags).Padded {
		p.PadLength = payload[0]
		index = 1
	}
	p.StreamID = binary.BigEndian.Uint32(payload[index:][:4]) & 0x7FFF // To remove the reserved bit
	p.Fragment = payload[:length-uint32(p.PadLength)][index+4:]
}

func (p PushPromisePayload) Bytes(flags IFlags) []byte {
	buffer := make([]byte, 4+len(p.Fragment)+int(p.PadLength))
	binary.BigEndian.PutUint32(buffer[:4], p.StreamID)
	copy(buffer[4:len(p.Fragment)], p.Fragment)
	if p.PadLength != 0 {
		buffer = append([]byte{p.PadLength}, buffer...)
	}
	return buffer
}

type PushPromise struct {
	Flags   PushPromiseFlags
	Payload PushPromisePayload
}

func CreatePushPromise(flags byte, payload []byte, payloadLength uint32) *PushPromise {
	push := &PushPromise{}
	push.Flags.ReadFlags(flags)
	push.Payload.ReadPayload(payload, payloadLength, &push.Flags)

	return push
}
