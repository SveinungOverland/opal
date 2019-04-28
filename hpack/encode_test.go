package hpack

import (
	"github.com/go-test/deep"
	"testing"
)

// ----- TESTS --------
func TestEncodeDecode(t *testing.T) {
	tests := getTestData01()
	for _, test := range tests {
		encodeDecodeTest(t, test)
	}

	tests = getTestData02()
	for _, test := range tests {
		encodeDecodeTest(t, test)
	}
}

func TestEncodeDecodeLongHeaders(t *testing.T) {
	longTests := getLongTestData()
	for _, test := range longTests {
		encodeDecodeTest(t, test)
	}
}

// ----- HELPERS ------

// Tests encode and decode
// If decode is tested and works, encode can be tested by trying to decode the encoded and
// check if equal
func encodeDecodeTest(t *testing.T, test hpackTest) {
	// Create encoder
	c1 := NewContext(256, 256)
	hfs := c1.Encode(test.expected)

	// Create decoder
	c2 := NewContext(256, 256)
	actual, err := c2.Decode(hfs)
	if err != nil {
		t.Error(err)
	}

	// Check if decoded headers are equal
	if diff := deep.Equal(actual, test.expected); diff != nil {
		t.Error(diff)
	}

	// Compare dynamic tables
	dynTabC1 := c1.EncoderDynamicTable()
	dynTabC2 := c2.DecoderDynamicTable()
	// Check if decoded headers are equal
	if diff := deep.Equal(dynTabC1, dynTabC2); diff != nil {
		t.Error(diff)
	}
}

// ----- TEST DATA ------

type encodeTest struct {
	test []*HeaderField
}

func getLongTestData() []hpackTest {
	return []hpackTest{
		{
			"",
			[]*HeaderField{
				hf("cuuuuuuuuuuuuuuuussssssssssssssssssssssssssstttttttttttttooooooooooooooooommmmmmmmmmmmmmm-------------keeeeeeeeeeeeeeeeyyyyyy", "cuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuusssssssssssstttttoooooooooooooooooooooooooooooommmmmmmmmmmmmm-vaaaaaaaaaaaaaaaaaaaaaaalllllllllllllllluuuuuuuuuuuuueeeeeeee"),
				hf("2348768gdh45ygxfg34rsdfg4tsr", "asdf543gfasdfh43hkuilusdfg35bfdfgh34sgfdhksdas345sdfgbnhteese345sdfgsdfhjsdfg45sdfgasdfasdfasdfaaaaaddfukkfds<fdDFHDFHDgfdhghdfghdfghdfgh"),
				hf("cookie", "token: eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJlYiI6InlnNTdSTDdjL2xxNjVLYmpxdWRaIiwianRpIjoiY2xhY2N0b2tfMDAwMDloWmZKcGthV1lPdnZ3RapplyIndexOrLength applyIndexOrLength applyIndexOrLength applyIndexOrLength applyIndexOrLength"),
			},
			nil,
		},
	}
}
