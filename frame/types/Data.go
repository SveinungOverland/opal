package types

import "io"

type DataFlags struct {
	EndStream bool
	Padded    bool
}

func (d DataFlags) ReadFlags(flags byte) {
	d.EndStream = (flags & 0x01) != 0x00
	d.Padded = (flags & 0x08) != 0x00
}

type DataPayload struct {
	Data []byte
}

func (d DataPayload) ReadPayload(r io.Reader, length uint32, flags IFlags) {
	padLength := make([]byte, 1)
	bytesToRead := length
	if flags.(DataFlags).Padded {
		r.Read(padLength)
		bytesToRead -= uint32(1 + uint8(padLength[0]))
	}

	payloadBuffer := make([]byte, bytesToRead)
	r.Read(payloadBuffer)

	d.Data = payloadBuffer
}

type Data struct {
	Flags   DataFlags
	Payload DataPayload
}

func CreateData(flags byte, payload io.Reader, payloadLength uint32) *Data {
	data := &Data{}
	data.Flags.ReadFlags(flags)
	data.Payload.ReadPayload(payload, payloadLength, data.Flags)

	return data
}
