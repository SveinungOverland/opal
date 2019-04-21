package main

import (
	"log"
	"opal"
)

func main() {
	srv, err := opal.NewTLSServer("./server.crt", "./server.key", nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(srv.Listen(8080))
}
