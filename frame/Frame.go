package frame

import (
	"encoding/binary"
	"io"
	"opal/frame/types"
)

// Create an enum for known types
//  server should ignore and discard unknown types

//Frame header is 9 octets

// Add known flags as booleans

/*  Type needs quite some work
go enums? If not need to implement
different structs for all the different
types of frame, with corresponding flags
*/

type Frame struct {
	// Remember to ignore the first bit of this field when reading
	ID      uint32
	Type    interface{} // Refactor this to enum
	Flags   types.IFlags
	Length  uint32
	Payload types.IPayload
}

func ReadFrame(r io.Reader) Frame {
	frame := Frame{}

	lengthBuffer := make([]byte, 3)
	r.Read(lengthBuffer)
	length := binary.BigEndian.Uint32(lengthBuffer)
	frame.Length = length

	typeFlagBuffer := make([]byte, 2)
	r.Read(typeFlagBuffer)

	switch frameType := typeFlagBuffer[0]; frameType {
	case 0x00: // Frame is of type Data  |  Carries request or response data
		data := types.CreateData(typeFlagBuffer[1], r, length)
		frame.Type = data
		frame.Flags = data.Flags
		frame.Payload = data.Payload
	case 0x01: // Frame is of type Headers  |  Carries request/response headers/trailers; can initiate a stream
		headers := types.CreateHeaders(typeFlagBuffer[1], r, length)
		frame.Type = headers
		frame.Flags = headers.Flags
		frame.Payload = headers.Payload
	case 0x02: // Frame is of type Priority  |  Indicates priority of a stream
	case 0x03: // Frame is of type RstStream  |  Terminates a stream
	case 0x04: // Frame is of type Settings  |  Defines parameters for the connection only
	case 0x05: // Frame is of type PushPromise  |  Signals peer for server push
	case 0x06: // Frame is of type Ping  |  Maintenance frame for checking RTT, connection, etc
	case 0x07: // Frame is of type GoAway  |  For shutting down a connection
	case 0x08: // Frame is of type WindowUpdate  |  Frame responsible for flow control adjustments
	case 0x09: // Frame is of type Continuation  |  Extends a HEADERS frame and can carry more headers
	}

	identifierBuffer := make([]byte, 4)
	r.Read(identifierBuffer)
	identifier := binary.BigEndian.Uint32(identifierBuffer) & 0x8000 // Bitwise and is used to remove the very first bit, as this is a reserved bit
	frame.ID = identifier

	payloadBuffer := make([]byte, length)
	r.Read(payloadBuffer)

	return frame
}
