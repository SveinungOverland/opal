package simpleHttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
)

type Response struct {
	Status int
	Body   []byte
	Header map[string]string
}

// ---- METHODS ------
func (res *Response) File(path string, contentType string) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	res.Body = file
	res.Header["Content-Type"] = contentType
}

func (res *Response) Html(path string, data interface{}) {
	temp, err := template.ParseFiles(path)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	var tpl bytes.Buffer
	if err := temp.Execute(&tpl, data); err != nil {
		return
	}
	res.Body = tpl.Bytes()
	res.Header["Content-Type"] = "text/html; charset=utf-8"
}

func (res *Response) Json(target interface{}) {
	res.Header["Content-Type"] = "application/json"
	json, _ := json.Marshal(target)

	res.Body = json
}

// ---- PRIVATE METHODS ----
func (r *Response) write(writer io.Writer) (int, error) {
	buf := make([]byte, 0, 1)
	w := bytes.NewBuffer(buf)

	// Write Response line
	w.WriteString("HTTP/1.1 ")
	w.WriteString(fmt.Sprintf("%d", r.Status))
	w.WriteString(fmt.Sprintf(" %s", http.StatusText(r.Status)))

	// Write headers
	w.WriteString("\n")
	for k, v := range r.Header {
		w.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}

	// Write content
	// Write content-length
	lengthOfBody := len(r.Body)
	if lengthOfBody > 0 {
		w.WriteString(fmt.Sprintf("Content-Length: %d\n", lengthOfBody))
		w.WriteString("\n")
		w.Write(r.Body)
	} else {
		w.WriteString("\n")
	}

	return writer.Write(w.Bytes())
}

// ---- HELPERS ------
func buildDefaultResponse() Response {
	res := Response{}
	res.Status = 200
	res.Body = make([]byte, 0)
	res.Header = make(map[string]string)

	res.Header["Content-Type"] = "text/plain"

	return res
}

func build404Response() Response {
	res := buildDefaultResponse()
	res.Html("./simpleHttp/templates/404.html", nil)
	res.Status = 404
	return res
}
