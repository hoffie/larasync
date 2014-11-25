package main

import (
	"log"
	"os"

	"github.com/larasync/lara/server"
)

// main is our service dispatcher.
func main() {
	if len(os.Args) <= 1 {
		log.Fatal("no action specified")
		os.Exit(1)
	}
	action := os.Args[1]
	switch action {
	case "server":
		s := server.New([]byte("FIXME-broken-hardcoded-secret")) //FIXME: config!
		log.Printf("Listening on :%d", server.DefaultPort)
		log.Fatal(s.ListenAndServe())
		return
	default:
		log.Fatal("unsupported action; possible actions: server")
		os.Exit(1)
	}
}
