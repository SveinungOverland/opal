package types

import "io"

type PingFlags struct {
	Ack bool
}

func (p PingFlags) ReadFlags(flags byte) {
	p.Ack = (flags & 0x01) != 0x00
}

type PingPayload struct {
	Data [8]byte
}

func (p PingPayload) ReadPayload(r io.Reader, length uint32, flags IFlags) {
	dataBuffer := make([]byte, 8)
	r.Read(dataBuffer)

	copy(p.Data[:], dataBuffer)
}

type Ping struct {
	Flags   PingFlags
	Payload PingPayload
}

func CreatePing(flags byte, payload io.Reader, payloadLength uint32) *Ping {
	ping := &Ping{}
	ping.Flags.ReadFlags(flags)
	ping.Payload.ReadPayload(payload, payloadLength, ping.Flags)

	return ping
}
