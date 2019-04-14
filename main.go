package main

import (
	"encoding/hex"
	"fmt"
	huff "opal/hpack/huffman"
)

func main() {
	fmt.Println("Building tree")
	root := huff.BuildTree()

	fmt.Println("Initializing")
	encodedHex := "a8eb10649cbf" //"f1e3c2e5f23a6ba0ab90f4ff"
	decoded, err := hex.DecodeString(encodedHex)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Decoding")
	result, err := huff.Decode(root, decoded)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(result))

	fmt.Println(13 & 1)

	test := " www-authenticate  "
	testBytes := []byte(test)
	fmt.Println(testBytes)
	encoded := huff.Encode(testBytes)
	fmt.Println(encoded)
	decoded, err = huff.Decode(root, encoded)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(decoded)
	fmt.Printf("\"%s\"\n", string(decoded))

}
