package types

import (
	"encoding/binary"
	"io"
)

type WindowUpdateFlags struct{}

func (w WindowUpdateFlags) ReadFlags(flags byte) {}

type WindowUpdatePayload struct {
	WindowSizeIncrement uint32
}

func (w WindowUpdatePayload) ReadPayload(r io.Reader, length uint32, flags IFlags) {
	sizeBuffer := make([]byte, 4)
	r.Read(sizeBuffer)

	w.WindowSizeIncrement = binary.BigEndian.Uint32(sizeBuffer) & 0x8000
}

type WindowUpdate struct {
	Flags   WindowUpdateFlags
	Payload WindowUpdatePayload
}

func CreateWindowUpdate(flags byte, payload io.Reader, payloadLength uint32) *WindowUpdate {
	window := &WindowUpdate{}
	window.Flags.ReadFlags(flags)
	window.Payload.ReadPayload(payload, payloadLength, window.Flags)

	return window
}
