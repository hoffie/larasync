package main

import (
	"log"
	"os"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/repository"
)

// main is our service dispatcher.
func main() {
	os.Exit(dispatch(os.Args[1:]))
}

func dispatch(args []string) int {
	if len(args) < 1 {
		log.Fatal("no action specified")
		return 1
	}
	action := args[0]
	if len(args) > 1 {
		flags.Parse(args[1:])
	}
	switch action {
	case "server":
		cfg := getServerConfig()
		rm, err := repository.NewManager(cfg.Repository.BasePath)
		if err != nil {
			log.Fatal("repository.Manager creation failure:", err)
		}
		s := api.New(*cfg.Signatures.AdminPubkeyBinary,
			cfg.Signatures.MaxAge, rm)
		log.Printf("Listening on %s", cfg.Server.Listen)
		log.Fatal(s.ListenAndServe())
		return 1
	default:
		log.Fatal("unsupported action; possible actions: server")
		return 1
	}
	return 0
}
