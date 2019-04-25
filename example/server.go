package main

import (
	"log"
	"opal"
	"opal/router"
	opalHttp "opal/http"
	_ "net/http/pprof"
	"net/http"
	"fmt"
)

func main() {
	go func() {
		fmt.Println("PPROF listening on :6060")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	srv, err := opal.NewTLSServer("./server.crt", "./server.key", nil)
	if err != nil {
		log.Fatal(err)
	}

	r := router.NewRouter("/test")
	r.Get("/", func(req *opalHttp.Request, res *opalHttp.Response) {
		res.Body = []byte("Hello World")
	})

	srv.Register(r)


	log.Fatal(srv.Listen(8080))
}
