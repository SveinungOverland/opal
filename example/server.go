package main

import (
	"log"
	"opal/core"
)

func main() {
	srv, err := core.NewTLSServer("./server.crt", "./server.key", nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(srv.Listen(8080))
}
