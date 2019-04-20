package types

import (
	"encoding/binary"
)

type WindowUpdateFlags struct{}

func (w WindowUpdateFlags) ReadFlags(flags byte) {}

func (w WindowUpdateFlags) Byte() (flags byte) {
	return
}

type WindowUpdatePayload struct {
	WindowSizeIncrement uint32
}

func (w *WindowUpdatePayload) ReadPayload(payload []byte, length uint32, flags IFlags) {
	w.WindowSizeIncrement = binary.BigEndian.Uint32(payload[:4]) & 0x7FFF
}

func (w WindowUpdatePayload) Bytes(flags IFlags) []byte {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, w.WindowSizeIncrement)
	return buffer
}

type WindowUpdate struct {
	Flags   WindowUpdateFlags
	Payload WindowUpdatePayload
}

func CreateWindowUpdate(flags byte, payload []byte, payloadLength uint32) *WindowUpdate {
	window := &WindowUpdate{}
	window.Flags.ReadFlags(flags)
	window.Payload.ReadPayload(payload, payloadLength, window.Flags)

	return window
}
