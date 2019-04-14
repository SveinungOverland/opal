package types

import "io"

// IFlags is an interface implemented by the different frame types so that the generic Frame struct can handle all the different frame types
type IFlags interface {
	ReadFlags(byte)
}

// IPayload is an interface implemented by the different frame types so that the generic Frame struct can handle all the different payload types
type IPayload interface {
	ReadPayload(r io.Reader, length uint32, flags IFlags)
}
