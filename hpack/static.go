package hpack

func getStaticHF(index uint32) *HeaderField {
	if index == 0 || index > uint32(len(staticTableEntries)) {
		return nil
	}
	return &staticTableEntries[index-1] // Index address space starts at 1, not 0 (sadly :( )
}

// Attempts to find corresponding headerfield in the static table. If a match in
// name is found, idx indicates the index (starts at 1), but perfectMatch indicates a match
// both in name and value.
func findStaticHF(hf *HeaderField) (idx uint32, perfectMatch bool) {
	perfectMatch = false

	for i, shf := range staticTableEntries {
		nameMatch, valueMatch := shf.equal(hf)
		if nameMatch && valueMatch {
			idx = uint32(i + 1) // Index address space starts at 1, not 0
			perfectMatch = true
			break // Stop search
		} else if nameMatch {
			idx = uint32(i + 1)
		}
	}
	return idx, perfectMatch
}

var staticTableEntries = [...]HeaderField{
	{Name: ":authority"},
	{Name: ":method", Value: "GET"},
	{Name: ":method", Value: "POST"},
	{Name: ":path", Value: "/"},
	{Name: ":path", Value: "/index.html"},
	{Name: ":scheme", Value: "http"},
	{Name: ":scheme", Value: "https"},
	{Name: ":status", Value: "200"},
	{Name: ":status", Value: "204"},
	{Name: ":status", Value: "206"},
	{Name: ":status", Value: "304"},
	{Name: ":status", Value: "400"},
	{Name: ":status", Value: "404"},
	{Name: ":status", Value: "500"},
	{Name: "accept-charset"},
	{Name: "accept-encoding", Value: "gzip, deflate"},
	{Name: "accept-language"},
	{Name: "accept-ranges"},
	{Name: "accept"},
	{Name: "access-control-allow-origin"},
	{Name: "age"},
	{Name: "allow"},
	{Name: "authorization"},
	{Name: "cache-control"},
	{Name: "content-disposition"},
	{Name: "content-encoding"},
	{Name: "content-language"},
	{Name: "content-length"},
	{Name: "content-location"},
	{Name: "content-range"},
	{Name: "content-type"},
	{Name: "cookie"},
	{Name: "date"},
	{Name: "etag"},
	{Name: "expect"},
	{Name: "expires"},
	{Name: "from"},
	{Name: "host"},
	{Name: "if-match"},
	{Name: "if-modified-since"},
	{Name: "if-none-match"},
	{Name: "if-range"},
	{Name: "if-unmodified-since"},
	{Name: "last-modified"},
	{Name: "link"},
	{Name: "location"},
	{Name: "max-forwards"},
	{Name: "proxy-authenticate"},
	{Name: "proxy-authorization"},
	{Name: "range"},
	{Name: "referer"},
	{Name: "refresh"},
	{Name: "retry-after"},
	{Name: "server"},
	{Name: "set-cookie"},
	{Name: "strict-transport-security"},
	{Name: "transfer-encoding"},
	{Name: "user-agent"},
	{Name: "vary"},
	{Name: "via"},
	{Name: "www-authenticate"},
}
