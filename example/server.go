package main

import (
	"log"
	"opal"
	"opal/router"
	"opal/http"
)

func main() {
	srv, err := opal.NewTLSServer("./server.crt", "./server.key", nil)
	if err != nil {
		log.Fatal(err)
	}

	r := router.NewRouter("/test")
	r.Get("/", func(req *http.Request, res *http.Response) {
		res.Body = []byte("Hello World")
	})

	r.Get("/site", func(req *http.Request, res *http.Response) {
		res.Body = []byte("<html><body><h4>Hello World! :D</h4><a href=\"\\\">Click here</a></body></html>")	
		res.Header["content-type"] = "text/html; charset=utf-8"
	})

	srv.Register(r)

	log.Fatal(srv.Listen(8080))
}
