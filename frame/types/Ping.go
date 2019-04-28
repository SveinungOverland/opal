package types

type PingFlags struct {
	Ack bool
}

func (p *PingFlags) ReadFlags(flags byte) {
	p.Ack = (flags & 0x1) != 0x0
}

func (p PingFlags) Byte() (flags byte) {
	if p.Ack {
		flags |= 0x1
	}
	return
}

type PingPayload struct {
	Data []byte
}

func (p *PingPayload) ReadPayload(payload []byte, length uint32, flags IFlags) {
	p.Data = payload
}

func (p PingPayload) Bytes(flags IFlags) []byte {
	return p.Data
}

type Ping struct {
	Flags   PingFlags
	Payload PingPayload
}

func CreatePing(flags byte, payload []byte, payloadLength uint32) *Ping {
	ping := &Ping{}
	ping.Flags.ReadFlags(flags)
	ping.Payload.ReadPayload(payload, payloadLength, &ping.Flags)

	return ping
}
