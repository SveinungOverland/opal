package hpack_test

import (
	"encoding/hex"
	"opal/hpack"
	"testing"

	"github.com/go-test/deep"
)

// TestDecode tests Decode.decode()
func TestDecode(t *testing.T) {
	testData := getTestData01()
	context := hpack.NewContext(256)
	testContextDecode(t, context, testData)
}

// TestDecode02 tests Decode.decode()
func TestDecode02(t *testing.T) {
	testData := getTestData02()
	context := hpack.NewContext(256)
	testContextDecode(t, context, testData)
}

// ----- HELPERS -------

func testContextDecode(t *testing.T, context *hpack.Context, testData []hpackTest) {
	for _, test := range testData {
		testBytes, _ := hex.DecodeString(test.hex)
		actual, err := context.Decode(testBytes)
		if err != nil {
			t.Errorf("[TestDeocde]: Error: %s", err.Error())
		}

		// Check if decoded headers are equal
		if diff := deep.Equal(actual, test.expected); diff != nil {
			t.Error(diff)
		}

		// Check if dynamic table is equal to expected
		dynTabHfs := context.DynamicTable()
		if diff := deep.Equal(dynTabHfs, test.expectedDynamicTable); diff != nil {
			t.Error(diff)
		}
	}
}

func hf(name string, value string) *hpack.HeaderField {
	return &hpack.HeaderField{name, value}
}

// --------- TEST DATA ------------

// Initialize test data
// All test data comes from RFC7541
// http://http2.github.io/http2-spec/compression.html#request.examples.with.huffman.coding
type hpackTest struct {
	hex                  string
	expected             []*hpack.HeaderField
	expectedDynamicTable []*hpack.HeaderField
}

func getTestData01() []hpackTest {
	return []hpackTest{
		// C.4 Request example
		{"828684418cf1e3c2e5f23a6ba0ab90f4ff", []*hpack.HeaderField{
			hf(":method", "GET"),
			hf(":scheme", "http"),
			hf(":path", "/"),
			hf(":authority", "www.example.com"),
		},
			[]*hpack.HeaderField{
				hf(":authority", "www.example.com"),
			},
		},

		{"828684be5886a8eb10649cbf", []*hpack.HeaderField{
			hf(":method", "GET"),
			hf(":scheme", "http"),
			hf(":path", "/"),
			hf(":authority", "www.example.com"),
			hf("cache-control", "no-cache"),
		},
			[]*hpack.HeaderField{
				hf("cache-control", "no-cache"),
				hf(":authority", "www.example.com"),
			},
		},
		{"828785bf408825a849e95ba97d7f8925a849e95bb8e8b4bf", []*hpack.HeaderField{
			hf(":method", "GET"),
			hf(":scheme", "https"),
			hf(":path", "/index.html"),
			hf(":authority", "www.example.com"),
			hf("custom-key", "custom-value"),
		},
			[]*hpack.HeaderField{
				hf("custom-key", "custom-value"),
				hf("cache-control", "no-cache"),
				hf(":authority", "www.example.com"),
			},
		},
	}
}

func getTestData02() []hpackTest {
	// C.6 Response example
	return []hpackTest{
		{"488264025885aec3771a4b6196d07abe941054d444a8200595040b8166e082a62d1bff6e919d29ad171863c78f0b97c8e9ae82ae43d3",
			[]*hpack.HeaderField{
				hf(":status", "302"),
				hf("cache-control", "private"),
				hf("date", "Mon, 21 Oct 2013 20:13:21 GMT"),
				hf("location", "https://www.example.com"),
			},
			[]*hpack.HeaderField{
				hf("location", "https://www.example.com"),
				hf("date", "Mon, 21 Oct 2013 20:13:21 GMT"),
				hf("cache-control", "private"),
				hf(":status", "302"),
			},
		},
		{"4883640effc1c0bf",
			[]*hpack.HeaderField{
				hf(":status", "307"),
				hf("cache-control", "private"),
				hf("date", "Mon, 21 Oct 2013 20:13:21 GMT"),
				hf("location", "https://www.example.com"),
			},
			[]*hpack.HeaderField{
				hf(":status", "307"),
				hf("location", "https://www.example.com"),
				hf("date", "Mon, 21 Oct 2013 20:13:21 GMT"),
				hf("cache-control", "private"),
			},
		},
		{"88c16196d07abe941054d444a8200595040b8166e084a62d1bffc05a839bd9ab77ad94e7821dd7f2e6c7b335dfdfcd5b3960d5af27087f3672c1ab270fb5291f9587316065c003ed4ee5b1063d5007",
			[]*hpack.HeaderField{
				hf(":status", "200"),
				hf("cache-control", "private"),
				hf("date", "Mon, 21 Oct 2013 20:13:22 GMT"),
				hf("location", "https://www.example.com"),
				hf("content-encoding", "gzip"),
				hf("set-cookie", "foo=ASDJKHQKBZXOQWEOPIUAXQWEOIU; max-age=3600; version=1"),
			},
			[]*hpack.HeaderField{
				hf("set-cookie", "foo=ASDJKHQKBZXOQWEOPIUAXQWEOIU; max-age=3600; version=1"),
				hf("content-encoding", "gzip"),
				hf("date", "Mon, 21 Oct 2013 20:13:22 GMT"),
			},
		},
	}
}
