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

func main() {
	context := hpack.NewContext(256)

	fmt.Println(context.DynamicTableString())

	testHex := "828684418cf1e3c2e5f23a6ba0ab90f4ff"
	Decode(context, testHex)
	testHex = "828684be5886a8eb10649cbf"
	Decode(context, testHex)
	testHex = "828785bf408825a849e95ba97d7f8925a849e95bb8e8b4bf"
	Decode(context, testHex)
	testHex = "4803333032580770726976617465611d4d6f6e2c203231204f637420323031332032303a31333a323120474d546e1768747470733a2f2f7777772e6578616d706c652e636f6d"
	Decode(context, testHex)
	testHex = "4803333037c1c0bf"
	Decode(context, testHex)

	context = hpack.NewContext(256)
	testHex = "488264025885aec3771a4b6196d07abe941054d444a8200595040b8166e082a62d1bff6e919d29ad171863c78f0b97c8e9ae82ae43d3"
	Decode(context, testHex)
	testHex = "4883640effc1c0bf"
	Decode(context, testHex)
	testHex = "88c16196d07abe941054d444a8200595040b8166e084a62d1bffc05a839bd9ab77ad94e7821dd7f2e6c7b335dfdfcd5b3960d5af27087f3672c1ab270fb5291f9587316065c003ed4ee5b1063d5007"
	Decode(context, testHex)
}
