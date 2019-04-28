package frame

import (
	"bytes"
	"github.com/SveinungOverland/opal/errors"
	"github.com/SveinungOverland/opal/frame/types"
	"reflect"
	"testing"
)

var testFrame = &Frame{
	ID:    0,
	Type:  PingType,
	Flags: &types.PingFlags{},
	Payload: &types.PingPayload{
		Data: make([]byte, 8),
	},
	Length: 8,
}

var testBytes = []byte{0, 0, 8, 6, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

func TestFrameToBytes(t *testing.T) {
	frameBytes := testFrame.ToBytes()
	if !bytes.Equal(frameBytes, testBytes) {
		t.Error("frame.ToBytes() did not encode properly")
	}
}

func TestFrameRead(t *testing.T) {
	frame, err := ReadFrame(bytes.NewReader(testBytes))
	if err != nil {
		t.Error("ReadFrame could not read frame")
	}
	if frame.ID != testFrame.ID {
		t.Error("Ids did not match")
	}
	if frame.Type != testFrame.Type {
		t.Error("Types did not match")
	}
	if frame.Length != testFrame.Length {
		t.Error("Lengths did not match")
	}
	if !reflect.DeepEqual(&frame.Flags, &testFrame.Flags) {
		t.Error("Flags did not match")
	}
	if !reflect.DeepEqual(&frame.Payload, &testFrame.Payload) {
		t.Error("Payload did not math")
	}
}

func TestNewErrorFrame(t *testing.T) {
	testErrorBytes := []byte{0, 0, 4, 3, 0, 0, 0, 0, 0, 0, 0, 0, 11}

	newErrorFrame := NewErrorFrame(0, errors.EnhanceYourCalm)

	if !bytes.Equal(newErrorFrame.ToBytes(), testErrorBytes) {
		t.Error("Bytes did not match")
	}
}
