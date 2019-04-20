package types

import (
	"encoding/binary"
	"fmt"
)

type SettingsFlags struct {
	Ack bool
}

func (s *SettingsFlags) ReadFlags(flags byte) {
	s.Ack = (flags & 0x1) != 0x0
}

func (s SettingsFlags) Byte() (flags byte) {
	if s.Ack {
		flags |= 0x1
	}
	return
}

type SettingsPayload struct {
	IDValuePair map[uint16]uint32
}

func (s *SettingsPayload) ReadPayload(payload []byte, length uint32, flags IFlags) {
	s.IDValuePair = make(map[uint16]uint32)
	fmt.Println(length / 6)
	for i := uint32(0); i < length/6; i++ {
		s.IDValuePair[binary.BigEndian.Uint16(payload[i*6:][:2])] = binary.BigEndian.Uint32(payload[i*6+2:][:4])
	}
	fmt.Printf("%+v\n", s.IDValuePair)
}

func (s SettingsPayload) Bytes(flags IFlags) []byte {
	buffer := make([]byte, len(s.IDValuePair)*6)
	bufferIndex := 0
	for key, value := range s.IDValuePair {
		binary.BigEndian.PutUint16(buffer[bufferIndex:bufferIndex+2], key)
		bufferIndex += 2
		binary.BigEndian.PutUint32(buffer[bufferIndex:bufferIndex+4], value)
		bufferIndex += 4
	}
	return buffer
}

type Settings struct {
	Flags   SettingsFlags
	Payload SettingsPayload
}

func CreateSettings(flags byte, payload []byte, payloadLength uint32) *Settings {
	settings := &Settings{}
	settings.Flags.ReadFlags(flags)
	settings.Payload.ReadPayload(payload, payloadLength, &settings.Flags)

	return settings
}
