package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"opal/hpack"
)

func Decode(context *hpack.Context, testHex string) {
	testDump, err := hex.DecodeString(testHex)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Decoding")
	hfs, err := context.Decode(testDump)
	if err != nil {
		fmt.Println(err)
	}
	if hfs == nil {
		fmt.Println("Result is nil")
		return
	}

	for i, hf := range hfs {
		fmt.Println(fmt.Sprintf("%d - %s", i+1, hf.String()))
	}

	fmt.Println("Dynamic Table")
	fmt.Println(context.DynamicTableString())
}

func hf(name string, value string) *hpack.HeaderField {
	return &hpack.HeaderField{name, value}
}

func main() {
	context := hpack.NewContext(256)

	test := []*hpack.HeaderField{
		hf(":method", "GET"),
		hf(":scheme", "http"),
		hf(":path", "/"),
		hf(":authority", "www.example.com"),
		hf("custom-key", "custom-value"),
		hf("cookie", "token: eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJlYiI6InlnNTdSTDdjL2xxNjVLYmpxdWRaIiwianRpIjoiY2xhY2N0b2tfMDAwMDloWmZKcGthV1lPdnZ3RapplyIndexOrLength applyIndexOrLength applyIndexOrLength applyIndexOrLength applyIndexOrLength"),
	}

	encoded, err := context.Encode(test)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(encoded)
	fmt.Println(context.DynamicTableString())

	c := hpack.NewContext(256)
	hfs, err := c.Decode(encoded)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i, hf := range hfs {
		fmt.Println(fmt.Sprintf("%d - %s", i+1, hf.String()))
	}

	fmt.Println(c.DynamicTableString())
}
