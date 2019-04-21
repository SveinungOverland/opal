package opal

type Request struct {
	Method     string
	RequestURI string
	RawQuery   string
	Params     map[string]string
	Header     map[string]string
	Body       []byte

	next bool   // Bool for deciding if next request can be handled
	Next func() // Changes the next-value to true
}
