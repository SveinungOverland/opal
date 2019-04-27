package main

import (
	"log"
	"opal"
	"opal/router"
	/* opalHttp "opal/http" */
	_ "net/http/pprof"
	"net/http"
	"fmt"
)

func main() {
	go func() {
		fmt.Println("PPROF listening on :6060, use /debug/pprof for overview and /debug/pprof/profile?seconds=20 for CPU-profiling (20s)")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	srv, err := opal.NewTLSServer("./server.crt", "./server.key", nil)
	if err != nil {
		log.Fatal(err)
	}

	r := router.NewRouter("/")
	/* r.Get("/", func(req *opalHttp.Request, res *opalHttp.Response) {
		res.Body = []byte("Hello World")
	})

	r.Get("/site", func(req *opalHttp.Request, res *opalHttp.Response) {
		res.Body = []byte("<html><head><link rel=\"stylesheet\" type=\"text/css\" href=\"css\\theme.css\"></head><body><h4>Hello World! :D</h4><a href=\"\\\">Click here</a></body></html>")	
		res.Header["content-type"] = "text/html; charset=utf-8"

		pushClient := opalHttp.NewPusher(req, res)
		pushClient.Push("/css/theme.css")
	}) */
	r.Static("/", "./build")

	fmt.Println(r.Root())

	/* r.Static("/css", "./css") */

	srv.Register(r)
	


	log.Fatal(srv.Listen(8080))
}
