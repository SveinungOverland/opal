package huff

import (
	"errors"
	"fmt"
)

type node struct {
	sym         int   // ASCII-value
	right, left *node // Children
}

func (n *node) String() string {
	return fmt.Sprintf("%d - %p %p\n", n.sym, n.right, n.left)
}

func createNode() *node {
	return &node{-1, nil, nil}
}

// BuildTree buildes the huffman tree based on the ASCII-table and returns the root node
func BuildTree() (root *node) {
	return buildTree(huffTable)
}

func buildTree(table *HuffTable) (root *node) {
	root = createNode()

	// Iterate through Huffman-table (every index is an ASCII-value)
	for i, huffCode := range table {
		curNode := root

		// Iterate through every bit
		j := huffCode.length
		for j > 0 {
			j--
			mask := uint32(1 << j) // Creates a sequence of bytes with only one 1. If length = 3 => 00001000

			// Check if corresponding bit at index j + 1 has a value of 1
			if huffCode.code&mask == mask {
				// The bit-value is 1, go to the right node
				nextNode := curNode.right
				if nextNode == nil { // right node does not exist
					nextNode = createNode()
					curNode.right = nextNode
				}
				curNode = nextNode
			} else {
				// The bit-value is 0, go to the left node
				nextNode := curNode.left
				if nextNode == nil { // left node does not exist
					nextNode = createNode()
					curNode.left = nextNode
				}
				curNode = nextNode
			}
		}

		// CurNode is now the leaf-node, set value (ASCII-value)
		curNode.sym = i
	}

	return root
}

// Decode decodes an array of huffman-encoded bytes
func Decode(root *node, data []byte) ([]byte, error) {
	if root == nil {
		return nil, errors.New("Root node can not be nil")
	}
	if len(data) == 0 {
		return data, nil
	}

	// Byte-array of decoded values
	var decoded []byte

	curNode := root
	for _, val := range data {

		// Reads every bit from left to right, and moves downwards from the root
		// If bit equals to 1, go to the right, else go left

		m := byte(128) // Mask is for getting the value of the bit, initial value is 1000 0000
		for m > 0 {
			// Check if bit-value is equal to 1
			if val&m == m {
				curNode = curNode.right
			} else {
				curNode = curNode.left
			}

			// Check if curNode is nil
			if curNode == nil {
				return decoded, errors.New("Invalid encoding")
			}

			// Check if leaf node
			if curNode.sym != -1 {
				decoded = append(decoded, byte(curNode.sym)) // Add symbol
				curNode = root                               // Start from the
			}
			m = m >> 1 // Move 1-value in mask to the right
		}
	}

	return decoded, nil
}

// Encode encodes an array of bytes to huffman encoded bytes
func Encode(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	encoded := make([]byte, 0)

	byteIndex := byte(7)  // remaining bits to set in value
	nextByte := byte(255) // A byte to store the encoded bits

	// For every byte
	for _, b := range data {

		huffCode := huffTable[b]
		curIndex := huffCode.length // The index of the huffman-code
		// Iterate through every bit in huff-code
		for curIndex > 0 {
			curIndex--

			// Check if bit is equal to 1
			mask := uint32(1 << curIndex)
			if huffCode.code&mask == mask {
				nextByte |= 1 << byteIndex // Sets bit at index 'remainingBits' to 1
			} else {
				nextByte &= ^(1 << byteIndex) // Clears bit at index 'remainingBits' (sets to 0)
			}

			// Check if nextByte is done
			if byteIndex == 0 {
				encoded = append(encoded, nextByte)
				byteIndex = 7
				nextByte = 255
			} else {
				byteIndex--
			}
		}

	}

	// Check for padding issues
	// RFC7541 5.2 says the last padding bits should be the most
	// significant bits of the huffcode corresponding to EOS (end-of-string) symbol.
	if byteIndex < 7 {
		// !NB - The most significan bits of the huffcode for EOS is always 1's. Therefore,
		// nextByte has a default value of 255 (1111 1111)

		/* eos := huffTable[256]
		diff := eos.length - remainingBits */
		encoded = append(encoded, nextByte)
	}

	return encoded
}
