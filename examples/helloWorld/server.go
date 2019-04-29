package main

import (
	"fmt"
	"github.com/SveinungOverland/opal"
	"github.com/SveinungOverland/opal/http"
	"github.com/SveinungOverland/opal/router"
	"log"
	"time"
)

func main() {
	// opal.NewTLSServer takes a third argument which is a error channel for connection errors, useful for debugging
	srv, err := opal.NewTLSServer("../server.crt", "../server.key", nil)
	if err != nil {
		panic(err)
	}

	r := router.NewRouter("/")

	r.Static("/assets", "./")

	r.Get("timenow", func(req *http.Request, res *http.Response) {
		res.String(200, time.Now().String())
	})

	r.Get("ping", func(req *http.Request, res *http.Response) {
		res.File("./pong.html")
		res.Push("/assets/pong.css")
	})

	r.Get("/api/:msg", func(req *http.Request, res *http.Response) {
		res.JSON(200, http.JSON{
			"hello": req.Param("msg"),
		})
	})

	// .Root() gives you a string displaying all the routes in the router
	fmt.Println(r.Root())

	srv.Register(r)

	log.Fatal(srv.Listen(5000))
}
