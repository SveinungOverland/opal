package huff_test

import (
	"encoding/hex"
	huff "opal/hpack/huffman"
	"testing"
)

// TestData recieved from RFC7541 - Section C.4
var testData = []struct {
	value, encodedHex string
}{
	{"www.example.com", "f1e3c2e5f23a6ba0ab90f4ff"},
	{"no-cache", "a8eb10649cbf"},
	{"custom-key", "25a849e95ba97d7f"},
	{"custom-value", "25a849e95bb8e8b4bf"},
	{"302", "6402"},
	{"private", "aec3771a4b"},
	{"Mon, 21 Oct 2013 20:13:21 GMT", "d07abe941054d444a8200595040b8166e082a62d1bff"},
	{"https://www.example.com", "9d29ad171863c78f0b97c8e9ae82ae43d3"},
	{"foo=ASDJKHQKBZXOQWEOPIUAXQWEOIU; max-age=3600; version=1", "94e7821dd7f2e6c7b335dfdfcd5b3960d5af27087f3672c1ab270fb5291f9587316065c003ed4ee5b1063d5007"},
}

func TestEncode(t *testing.T) {
	for _, test := range testData {
		expected := test.encodedHex
		actualBytes := huff.Encode([]byte(test.value))
		actual := hex.EncodeToString(actualBytes)
		assertEqual(t, actual, expected)
	}
}

func TestDecode(t *testing.T) {
	for _, test := range testData {
		expected := test.value
		bytes, _ := hex.DecodeString(test.encodedHex)
		actual, _ := huff.Decode(bytes)
		assertEqual(t, string(actual), expected)
	}
}

func TestEncodeAndDecode(t *testing.T) {
	for _, test := range testData {
		expected := []byte(test.value)
		encoded := huff.Encode(expected)
		decoded, _ := huff.Decode(encoded)
		assertEqual(t, string(decoded), string(expected))
	}
}

func assertEqual(t *testing.T, actual string, expected string) {
	if actual != expected {
		t.Errorf("Expected: %s, got %s", expected, actual)
	}
}
