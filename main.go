package main

import (
	"fmt"
	"opal/router"
)

func handler(req interface{} /* *Request */, res interface{} /* *Response */) {

}

func main() {
	r := router.NewRouter("/")
	r.Get("/aaa/bbb/ccc", handler)
	r.Post("/aaa/bbb/dddd", handler)
	r.Get("/aaa/bbb/dddd", handler)
	r.Get("/aaa/BBc/", handler)
	r.Put("/aaa/:id", handler)
	r.Static("/image", "./public")
	r.Post("/123/23/path", handler)
	r.Put("/:lat/:lng/path/", handler)

	root := r.Root()

	fmt.Println(root.String())

	match, route, params, fh := router.Search(root, "/123/23/path")
	fmt.Println(match)
	fmt.Println(route)
	fmt.Println(params)
	fmt.Println(fh)

	if fh != nil {
		file, err := fh.ReadFile()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(file))
	}
}
