package http

import (
	"encoding/json"
	"io/ioutil"
	"html/template"
	"bytes"
	"fmt"
)

type Response struct {
	Status uint16
	Body   []byte
	Header map[string]string


	// RFC 7540 - Section 8.2 - Server Push
	req *Request // Requests corresponding to the Response. Is used for Server Push
	pushService *pushService // A pushService, storing all the push requests from Response.Push(path string)
}

// ----- BODY MODIFIERS -----

// JSON sets body to the given content in json-format 
func (res *Response) JSON(target interface{}) {
	res.Header["content-type"] = "application/json"
	json, _ := json.Marshal(target)

	res.Body = json
}

// File reads and sets the file
func (res *Response) File(path string, contentType string) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	res.Body = file
	res.Header["content-type"] = contentType
}

// HTML reads an html-file and builds a html-response. Includes support for templates.
func (res *Response) HTML(path string, templateData interface{}) {
	temp, err := template.ParseFiles(path)
	if err != nil {
		fmt.Println("Error parsing html file: ", err)
		return
	}

	var tpl bytes.Buffer
	if err := temp.Execute(&tpl, templateData); err != nil {
		return
	}
	res.Body = tpl.Bytes()
	res.Header["content-type"] = "text/html; charset=utf-8"
}

// ----- STATUS SETTERS ------

// Ok sets response status-code to 200
func (res *Response) Ok() {
	res.Status = 200
}

// Created sets response status-code to 201
func (res *Response) Created() {
	res.Status = 201
}

// BadRequest sets response status-code to 403
func (res *Response) BadRequest() {
	res.Status = 400
}

// Unauthorized sets response status-code to 401
func (res *Response) Unauthorized() {
	res.Status = 401
}

// Forbidden sets response status-code to 403
func (res *Response) Forbidden() {
	res.Status = 403
}

// NotFound sets response status-code to 404
func (res *Response) NotFound() {
	res.Status = 404
}


// ------ SERVER PUSH ---------

// Push performs a push server
func (res *Response) Push(path string) {
	// Check if pushService is initialized
	if res.pushService == nil {
		res.pushService = newPushService(res)
	}

	res.pushService.Push(path)
}

// PushRequests returns all initialized Push-Requests
func (res *Response) PushRequests() []*Request {
	if res.pushService == nil {
		return []*Request{}
	}
	return res.pushService.pushRequests
}

// ------ RESPONSE BUILDERS ------

// NewResponse creates a new response
func NewResponse(req *Request) *Response {
	res := &Response{
		Status: 200,
		Body:   make([]byte, 0),
		Header: make(map[string]string),
		req: req,
	}
	res.Header["content-type"] = "text/plain; charset=utf-8"
	return res
}
