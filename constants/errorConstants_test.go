package constants

import (
	"testing"
)

func TestNoError(t *testing.T) {
	if NoError != 0 {
		t.Fatal("NoError has wrong error code")
	}
}

func TestProtocolError(t *testing.T) {
	if ProtocolError != 1 {
		t.Fatal("ProtocolError has wrong error code")
	}
}

func TestInternalError(t *testing.T) {
	if InternalError != 2 {
		t.Fatal("InternalError has wrong error code")
	}
}

func TestFlowControlError(t *testing.T) {
	if FlowControlError != 3 {
		t.Fatal("FlowControlError has wrong error code")
	}
}

func TestSettingsTimeout(t *testing.T) {
	if SettingsTimeout != 4 {
		t.Fatal("SettingsTimeout has wrong error code")
	}
}

func TestStreamClosed(t *testing.T) {
	if StreamClosed != 5 {
		t.Fatal("StreamClosed has wrong error code")
	}
}

func TestFrameSizeError(t *testing.T) {
	if FrameSizeError != 6 {
		t.Fatal("FrameSizeError has wrong error code")
	}
}

func TestRefusedStream(t *testing.T) {
	if RefusedStream != 7 {
		t.Fatal("RefusedStream has wrong error code")
	}
}

func TestCancel(t *testing.T) {
	if Cancel != 8 {
		t.Fatal("Cancel has wrong error code")
	}
}

func TestCompressionError(t *testing.T) {
	if CompressionError != 9 {
		t.Fatal("CompressionError has wrong error code")
	}
}

func TestConnectError(t *testing.T) {
	if ConnectError != 10 {
		t.Fatal("ConnectError has wrong error code")
	}
}

func TestEnhanceYourCalm(t *testing.T) {
	if EnhanceYourCalm != 11 {
		t.Fatal("EnhanceYourCalm has wrong error code")
	}
}

func TestInadequateSecurity(t *testing.T) {
	if InadequateSecurity != 12 {
		t.Fatal("InadequateSecurity has wrong error code")
	}
}

func TestHTTP11Required(t *testing.T) {
	if HTTP11Required != 13 {
		t.Fatal("HTTP11Required has wrong error code")
	}
}
