package main

import (
	"log"
	"os"

	"github.com/hoffie/larasync/api"
)

// main is our service dispatcher.
func main() {
	if len(os.Args) <= 1 {
		log.Fatal("no action specified")
		os.Exit(1)
	}
	action := os.Args[1]
	if len(os.Args) > 1 {
		flags.Parse(os.Args[2:])
	}
	switch action {
	case "server":
		cfg := getServerConfig()
		s := api.New([]byte(cfg.Signatures.AdminSecret), cfg.Signatures.MaxAge)
		log.Printf("Listening on %s", cfg.Server.Listen)
		log.Fatal(s.ListenAndServe())
		return
	default:
		log.Fatal("unsupported action; possible actions: server")
		os.Exit(1)
	}
}
