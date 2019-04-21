package types

import (
	"encoding/binary"
)

type HeadersFlags struct {
	EndStream  bool
	EndHeaders bool
	Padded     bool
	Priority   bool
}

func (h *HeadersFlags) ReadFlags(flags byte) {
	h.EndStream = (flags & 0x1) != 0x0
	h.EndHeaders = (flags & 0x4) != 0x0
	h.Padded = (flags & 0x8) != 0x0
	h.Priority = (flags & 0x20) != 0x00
}

func (h HeadersFlags) Byte() (flags byte) {
	if h.EndStream {
		flags |= 0x1
	}
	if h.EndHeaders {
		flags |= 0x4
	}
	if h.Padded {
		flags |= 0x8
	}
	if h.Priority {
		flags |= 0x20
	}
	return
}

type HeadersPayload struct {
	StreamExclusive  bool   // Can only be set true if Priority flag is set
	StreamDependency uint32 // Can only be present if Priority flag  is set
	PriorityWeight   byte   // Can only be present if Priority flag is set
	Fragment         []byte
	PadLength        byte
}

func (h *HeadersPayload) ReadPayload(payload []byte, length uint32, flags IFlags) {
	index := 0
	if flags.(*HeadersFlags).Padded {
		h.PadLength = payload[0]
		index = 1
	}
	if flags.(*HeadersFlags).Priority {
		h.StreamDependency = binary.BigEndian.Uint32(payload[index:][:4]) & 0x7FFFFFFF
		h.StreamExclusive = payload[index]&0x80 != 0x00
		h.PriorityWeight = payload[index+4]
		index += 5
	}
	h.Fragment = payload[:length-uint32(h.PadLength)][index:]
}

func (h HeadersPayload) Bytes(flags IFlags) []byte {
	buffer := make([]byte, len(h.Fragment)+int(h.PadLength))

	copy(buffer[:len(h.Fragment)], h.Fragment)

	if flags.(*HeadersFlags).Priority {
		priBuffer := make([]byte, 5)
		binary.BigEndian.PutUint32(priBuffer[:4], h.StreamDependency)
		priBuffer[4] = h.PriorityWeight
		if h.StreamExclusive {
			priBuffer[0] |= 0x8
		}
		buffer = append(priBuffer, buffer...)
	}

	if h.PadLength != 0 {
		buffer = append([]byte{h.PadLength}, buffer...)
	}

	return buffer
}

type Headers struct {
	Flags   HeadersFlags
	Payload HeadersPayload
}

func CreateHeaders(flags byte, payload []byte, payloadLength uint32) *Headers {
	headers := &Headers{}
	headers.Flags.ReadFlags(flags)
	headers.Payload.ReadPayload(payload, payloadLength, &headers.Flags)

	return headers
}
