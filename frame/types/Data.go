package types

type DataFlags struct {
	EndStream bool
	Padded    bool
}

func (d *DataFlags) ReadFlags(flags byte) {
	d.EndStream = (flags & 0x1) != 0x0
	d.Padded = (flags & 0x8) != 0x0
}

func (d DataFlags) Byte() (flags byte) {
	if d.EndStream {
		flags |= 0x1
	}
	if d.Padded {
		flags |= 0x8
	}
	return
}

type DataPayload struct {
	Data []byte
}

func (d *DataPayload) ReadPayload(payload []byte, length uint32, flags IFlags) {
	if flags.(*DataFlags).Padded {
		d.Data = payload[:length-uint32(payload[0])][1:]
		return
	}
	d.Data = payload
}

func (d DataPayload) Bytes(flags IFlags) []byte {
	return d.Data
}

type Data struct {
	Flags   DataFlags
	Payload DataPayload
}

func CreateData(flags byte, payload []byte, payloadLength uint32) *Data {
	data := &Data{}
	data.Flags.ReadFlags(flags)
	data.Payload.ReadPayload(payload, payloadLength, &data.Flags)

	return data
}
