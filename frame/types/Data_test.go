package types

import (
	"testing"
)

var testFlagsByte = byte(0x1)
var testFlagsStruct = DataFlags{
	EndStream: true,
}

func TestBytesToFlag(t *testing.T) {
	flags := DataFlags{}
	flags.ReadFlags(testFlagsByte)
	if flags.EndStream != true {
		t.Error("Flags were not read correctly")
	}
}
