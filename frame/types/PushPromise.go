package types

import (
	"encoding/binary"
	"io"
)

type PushPromiseFlags struct {
	EndHeaders bool
	Padded     bool
}

func (p PushPromiseFlags) ReadFlags(flags byte) {
	p.EndHeaders = (flags & 0x04) != 0x00
	p.Padded = (flags & 0x08) != 0x00
}

type PushPromisePayload struct {
	StreamID uint32
	Fragment []byte
}

func (p PushPromisePayload) ReadPayload(r io.Reader, length uint32, flags IFlags) {
	padLength := make([]byte, 1)
	bytesToRead := length
	if flags.(PushPromiseFlags).Padded {
		r.Read(padLength)
		bytesToRead -= uint32(1 + uint8(padLength[0]))
	}
	bytesToRead -= 4

	streamIDBuffer := make([]byte, 4)
	r.Read(streamIDBuffer)
	p.StreamID = binary.BigEndian.Uint32(streamIDBuffer) & 0x8000

	fragmentBuffer := make([]byte, bytesToRead)
	r.Read(fragmentBuffer)

	p.Fragment = fragmentBuffer
}

type PushPromise struct {
	Flags   PushPromiseFlags
	Payload PushPromisePayload
}

func CreatePushPromise(flags byte, payload io.Reader, payloadLength uint32) *PushPromise {
	push := &PushPromise{}
	push.Flags.ReadFlags(flags)
	push.Payload.ReadPayload(payload, payloadLength, push.Flags)

	return push
}
