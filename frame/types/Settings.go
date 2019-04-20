package types

import (
	"encoding/binary"
	"fmt"
	"io"
)

type SettingsFlags struct {
	Ack bool
}

func (s SettingsFlags) ReadFlags(flags byte) {
	s.Ack = (flags & 0x1) != 0x0
}

type SettingsPayload struct {
	idValuePair map[uint16]uint32
}

func (s SettingsPayload) ReadPayload(r io.Reader, length uint32, flags IFlags) {
	s.idValuePair = make(map[uint16]uint32)
	fmt.Println(length / 6)
	idBuffer := make([]byte, 2)
	valueBuffer := make([]byte, 4)
	for i := uint32(0); i < length/6; i++ {
		r.Read(idBuffer)
		r.Read(valueBuffer)

		fmt.Println(idBuffer, valueBuffer)
		s.idValuePair[binary.BigEndian.Uint16(idBuffer)] = binary.BigEndian.Uint32(valueBuffer)
	}
	fmt.Printf("%+v\n", s.idValuePair)
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
