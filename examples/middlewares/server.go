package main

import (
	"github.com/SveinungOverland/opal"
	"github.com/SveinungOverland/opal/http"
	"github.com/SveinungOverland/opal/router"
	"log"
)

func cors() router.HandleFunc {
	return func(req *http.Request, res *http.Response) {
		res.Header["Access-Control-Allow-Origin"] = "*"
		res.Header["Access-Control-Allow-Headers"] = "Origin, X-Requested-With, Content-Type, Accept, Authorization, X-CSRF-Token"
		if req.Method == "OPTIONS" {
			res.Header["Access-Control-Allow-Methods"] = "GET, POST, PATCH, DELETE, PUT"
			res.Status = 204
			req.Finish()
		}
	}
}

func isAuthenticated(req *http.Request, res *http.Response) {
	token := req.Query("token")
	if token != "1234" {
		res.Unauthorized()
		req.Finish()
	}
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	// Initialize new server
	srv, err := opal.NewTLSServer("../cert/server.crt", "../cert/server.key")
	handleError(err)

	// Add CORS
	srv.Use(cors())

	// Initialize routes
	r := router.NewRouter("/")

	r.Post("/private", isAuthenticated, func(req *http.Request, res *http.Response) {
		res.String(200, "I am authenticated! :D")
	})

	srv.Register(r)
	log.Fatal(srv.Listen(8080))
}
