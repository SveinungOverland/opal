package main

import (
	"github.com/SveinungOverland/opal"
	"github.com/SveinungOverland/opal/http"
	"github.com/SveinungOverland/opal/router"
	"log"
)

func main() {

	// Initialize new server
	srv, err := opal.NewTLSServer("../cert/server.crt", "../cert/server.key")
	if err != nil {
		log.Fatal(err)
	}
	r := router.NewRouter("/")

	// Endpoint returning HTML and pushing extra resources we know the client will need.
	r.Get("/site", func(req *http.Request, res *http.Response) {
		res.Body = []byte("<html><head><link rel=\"stylesheet\" type=\"text/css\" href=\"css\\theme.css\"></head><body><div id=\"main\"><h4>Hello World! :D</h4><a href=\"\\\">Click here</a></div><img src=\"css\\logo.png\" /><script src=\"css\\app.js\"></script></body></html>")
		res.Header["content-type"] = "text/html; charset=utf-8"

		res.Push("/css/theme.css")
		res.Push("/css/app.js")
		res.Push("/css/logo.png")
	})
	// Making all files in ./css accessable
	r.Static("/css", "./css")

	srv.Register(r)
	log.Fatal(srv.Listen(8080))
}
