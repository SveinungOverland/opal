package main

import (
	"fmt"
	"log"
	"opal"
	opalHttp "opal/http"
	"opal/router"
	"os"
	"runtime/pprof"
	"time"
)

func main() {

	file, _ := os.Create("profile")
	pprof.StartCPUProfile(file)

	go func() {
		defer pprof.StopCPUProfile()
		time.Sleep(30 * time.Second)
	}()

	srv, err := opal.NewTLSServer("./server.crt", "./server.key", nil)
	if err != nil {
		log.Fatal(err)
	}

	r := router.NewRouter("/")
	r.Get("/", func(req *opalHttp.Request, res *opalHttp.Response) {
		res.Body = []byte("Hello World")
	})

	r.Get("/site", func(req *opalHttp.Request, res *opalHttp.Response) {
		res.Body = []byte("<html><head><link rel=\"stylesheet\" type=\"text/css\" href=\"css\\theme.css\"></head><body><div id=\"main\"><h4>Hello World! :D</h4><a href=\"\\\">Click here</a></div><img src=\"css\\logo.png\" /><script src=\"css\\app.js\"></script></body></html>")
		res.Header["content-type"] = "text/html; charset=utf-8"

		res.Push("/css/theme.css")
		res.Push("/css/app.js")
		res.Push("/css/logo.png")
	})
	r.Static("/css", "./css")

	/* r.Static("/", "./build") */

	fmt.Println(r.Root())

	srv.Register(r)

	log.Fatal(srv.Listen(8080))
}
