package frame

import (
	"encoding/binary"
	// "fmt"
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

const (
	DataType byte = iota
	HeadersType
	PriorityType
	RstStreamType
	SettingsType
	PushPromiseType
	PingType
	GoAwayType
	WindowUpdateType
	ContinuationType
)

// Frame describes the data that is needed to receive and send a frame
type Frame struct {
	// Remember to ignore the first bit of this field when reading
	ID      uint32
	Type    byte
	Flags   types.IFlags
	Length  uint32
	Payload types.IPayload
}

// ReadFrame takes a reader and returns a frame with type
func ReadFrame(r io.Reader) (Frame, error) {
	// fmt.Println("Reading frame")
	frame := Frame{}

	lengthBuffer := make([]byte, 3)
	if _, err := r.Read(lengthBuffer); err != nil {
		return frame, err
	}
	length := binary.BigEndian.Uint32(append([]byte{0}, lengthBuffer...))
	// fmt.Println("Read frame length equal to", length)
	frame.Length = length

	typeFlagBuffer := make([]byte, 2)
	r.Read(typeFlagBuffer)

	identifierBuffer := make([]byte, 4)
	r.Read(identifierBuffer)
	identifier := binary.BigEndian.Uint32(identifierBuffer) & 0x7FFFFFFF // Bitwise 'and' is used to remove the very first bit, as this is a reserved bit
	frame.ID = identifier

	payloadBuffer := make([]byte, length)
	r.Read(payloadBuffer)

	// Handle frame payload dependent on frame type
	switch frameType := typeFlagBuffer[0]; frameType {
	case DataType: // Frame is of type Data  |  Carries request or response data
		data := types.CreateData(typeFlagBuffer[1], payloadBuffer, length)
		frame.Type = DataType
		frame.Flags = &data.Flags
		frame.Payload = &data.Payload
	case HeadersType: // Frame is of type Headers  |  Carries request/response headers/trailers; can initiate a stream
		headers := types.CreateHeaders(typeFlagBuffer[1], payloadBuffer, length)
		frame.Type = HeadersType
		frame.Flags = &headers.Flags
		frame.Payload = &headers.Payload
	case PriorityType: // Frame is of type Priority  |  Indicates priority of a stream
		priority := types.CreatePriority(typeFlagBuffer[1], payloadBuffer, length)
		frame.Type = PriorityType
		frame.Flags = &priority.Flags
		frame.Payload = &priority.Payload
	case RstStreamType: // Frame is of type RstStream  |  Terminates a stream
		rstStream := types.CreateRstStream(typeFlagBuffer[1], payloadBuffer, length)
		frame.Type = RstStreamType
		frame.Flags = &rstStream.Flags
		frame.Payload = &rstStream.Payload
	case SettingsType: // Frame is of type Settings  |  Defines parameters for the connection only
		settings := types.CreateSettings(typeFlagBuffer[1], payloadBuffer, length)
		frame.Type = SettingsType
		// fmt.Printf("%+v\n", settings.Flags)
		frame.Flags = &settings.Flags
		frame.Payload = &settings.Payload
	case PushPromiseType: // Frame is of type PushPromise  |  Signals peer for server push
		pushPromise := types.CreatePushPromise(typeFlagBuffer[1], payloadBuffer, length)
		frame.Type = PushPromiseType
		frame.Flags = &pushPromise.Flags
		frame.Payload = &pushPromise.Payload
	case PingType: // Frame is of type Ping  |  Maintenance frame for checking RTT, connection, etc
		ping := types.CreatePing(typeFlagBuffer[1], payloadBuffer, length)
		frame.Type = PingType
		frame.Flags = &ping.Flags
		frame.Payload = &ping.Payload
	case GoAwayType: // Frame is of type GoAway  |  For shutting down a connection
		goAway := types.CreateGoAway(typeFlagBuffer[1], payloadBuffer, length)
		frame.Type = GoAwayType
		frame.Flags = &goAway.Flags
		frame.Payload = &goAway.Payload
	case WindowUpdateType: // Frame is of type WindowUpdate  |  Frame responsible for flow control adjustments
		windowUpdate := types.CreateWindowUpdate(typeFlagBuffer[1], payloadBuffer, length)
		frame.Type = WindowUpdateType
		frame.Flags = &windowUpdate.Flags
		frame.Payload = &windowUpdate.Payload
	case ContinuationType: // Frame is of type Continuation  |  Extends a HEADERS frame and can carry more headers
		continuation := types.CreateContinuation(typeFlagBuffer[1], payloadBuffer, length)
		frame.Type = ContinuationType
		frame.Flags = &continuation.Flags
		frame.Payload = &continuation.Payload
	}

	return frame, nil
}

// ToBytes turnes a frame into sendable bytes
func (f *Frame) ToBytes() []byte {
	frameHeader := make([]byte, 9)

	lengthBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuffer, f.Length)
	copy(frameHeader[:3], lengthBuffer[1:])

	frameHeader[3] = f.Type
	frameHeader[4] = f.Flags.Byte()

	binary.BigEndian.PutUint32(frameHeader[5:], f.ID)

	if f.Length > 0 {
		return append(frameHeader, f.Payload.Bytes(f.Flags)...)
	}
	return frameHeader
}

// NewErrorFrame is a helper function for easily creating RstStream frames that occour on various errors
func NewErrorFrame(streamID, errorCode uint32) *Frame {
	return &Frame{
		ID: streamID,
		Type: RstStreamType,
		Flags: &types.RstStreamFlags{},
		Payload: &types.RstStreamPayload{
			ErrorCode: errorCode,
		},
		Length: 4,
	}
}
