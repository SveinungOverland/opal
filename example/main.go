package main

import (
	"fmt"
	"opal/router"
)

func handler(req interface{} /* *Request */, res interface{} /* *Response */) {
	fmt.Println("My cool and custom handler! :D")
}

func main() {
	mainRoot := router.NewRoot()

	r := router.NewRouter("/user/university/connection")
	r.Get("/aaa/bbb/ccc", handler)
	r.Post("/aaa/bbb/dddd", handler)
	r.Get("/aaa/bbb/dddd", handler)
	r.Get("/aaa/BBc/", handler)
	r.Put("/aaa/:id", handler)
	r.Static("/image", "./public")
	r.Post("/123/23/path", handler, handler, handler)
	r.Put("/:lat/:lng/path/", handler)

	mainRoot.AppendRouter(r)

	root := mainRoot

	fmt.Println(root.String())

	match, route, params, fh := root.Search("/123/23/path")
	fmt.Println(match)
	fmt.Println(route)
	fmt.Println(params)
	fmt.Println(fh)

	// Run handlers if a match was found
	if match {
		runHandlers := func(method string) {
			handlers := route.GetHandlers(method)
			for _, handler := range handlers {
				handler(nil, nil)
			}
		}
		runHandlers("GET")
		runHandlers("POST")
		runHandlers("PUT")
	}

	// If a filehandler was provided
	if fh != nil {
		file, err := fh.ReadFile()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(file))
	}
}
