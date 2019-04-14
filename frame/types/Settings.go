package types

import (
	"encoding/binary"
	"io"
)

type SettingsFlags struct {
	Ack bool
}

func (s SettingsFlags) ReadFlags(flags byte) {
	s.Ack = (flags & 0x1) != 0x0
}

type SettingsPayload struct {
	ID    uint16
	Value uint32
}

func (s SettingsPayload) ReadPayload(r io.Reader, length uint32, flags IFlags) {
	idBuffer := make([]byte, 2)
	r.Read(idBuffer)
	s.ID = binary.BigEndian.Uint16(idBuffer)

	valueBuffer := make([]byte, 4)
	r.Read(valueBuffer)
	s.Value = binary.BigEndian.Uint32(valueBuffer)
}

type Settings struct {
	Flags   SettingsFlags
	Payload SettingsPayload
}

func CreateSettings(flags byte, payload io.Reader, payloadLength uint32) *Settings {
	settings := &Settings{}
	settings.Flags.ReadFlags(flags)
	settings.Payload.ReadPayload(payload, payloadLength, settings.Flags)

	return settings
}
