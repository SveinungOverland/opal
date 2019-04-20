package main

import (
	"log"
	"opal/core"
)

func main() {
	srv, _ := core.NewTLSServer("./server.crt", "./server.key", nil)

	log.Fatal(srv.Listen(8080))
}
