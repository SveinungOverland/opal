package main

import (
	"log"
	"opal/core"
)

func main() {
	srv, _ := core.NewTLSServer("./server.crt", "./server.key")

	log.Fatal(srv.Listen(8080))
}
