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

	srv.Register(r)

	log.Fatal(srv.Listen(8080))
}
