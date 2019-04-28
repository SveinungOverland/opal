package types

import (
	"bytes"
	"reflect"
	"testing"
)

var testFlagsByte = byte(0x9)
var testFlagsStruct = DataFlags{
	EndStream: true,
	Padded:    true,
}
var testPayloadBytes = append([]byte{1}, []byte("Hello World ")...) // Space is added to the end to simulate padding
var testPayloadStruct = DataPayload{
	Data: []byte("Hello World"),
}

func TestDataBytesToFlag(t *testing.T) {
	flags := DataFlags{}
	flags.ReadFlags(testFlagsByte)
	if flags.EndStream != true {
		t.Error("Flag: EndStream was not correctly read")
	}
	if flags.Padded != true {
		t.Error("Flag: Padded was not correctly read")
	}
}

func TestDataFlagStructToBytes(t *testing.T) {
	testBytes := testFlagsStruct.Byte()
	if testBytes != testFlagsByte {
		t.Error("Flag was not correctly translated to byte")
	}
}

func TestDataReadPayload(t *testing.T) {
	payload := DataPayload{}
	payload.ReadPayload(testPayloadBytes, uint32(len(testPayloadBytes)), &testFlagsStruct)
	if !bytes.Equal(payload.Bytes(&testFlagsStruct), testPayloadStruct.Data) {
		t.Error("Payload was not read correctly")
	}
	payload2 := DataPayload{}
	payload2.ReadPayload(testPayloadBytes, uint32(len(testPayloadBytes)), &DataFlags{})
	if !bytes.Equal(payload2.Bytes(&DataFlags{}), testPayloadBytes) {
		t.Error("Reading payload with no padding didn't work correctly")
	}
}

func TestCreateData(t *testing.T) {
	data := CreateData(testFlagsByte, testPayloadBytes, uint32(len(testPayloadBytes)))
	if !reflect.DeepEqual(data.Flags, testFlagsStruct) {
		t.Error("CreateData did not return flags correctly")
	}
	if !reflect.DeepEqual(data.Payload, testPayloadStruct) {
		t.Error("CreateData did not return payload correctly")
	}
}
