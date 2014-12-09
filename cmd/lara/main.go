package main

import (
	"os"
	"fmt"

	"github.com/inconshreveable/log15"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/repository"
)

var log = log15.New("module", "main")

// main is our service dispatcher.
func main() {
	os.Exit(dispatch(os.Args[1:]))
}

func dispatch(args []string) int {
	if len(args) < 1 {
		fmt.Fprint(os.Stderr, "no action specified\n")
		return 1
	}
	action := args[0]
	if len(args) > 1 {
		flags.Parse(args[1:])
	}
	switch action {
	case "server":
		return serverAction()
	default:
		return defaultAction()
	}
	return 0
}

func setupLogging() {
	handler := log15.StreamHandler(os.Stderr, log15.LogfmtFormat())
	log.SetHandler(handler)
	repository.Log.SetHandler(handler)
	api.Log.SetHandler(handler)
}

func serverAction() int {
	setupLogging()
	cfg := getServerConfig()
	rm, err := repository.NewManager(cfg.Repository.BasePath)
	if err != nil {
		log.Error("repository.Manager creation failure", log15.Ctx{"error": err})
		return 1
	}
	s := api.New(*cfg.Signatures.AdminPubkeyBinary,
		cfg.Signatures.MaxAge, rm)
	log.Info("Listening", log15.Ctx{"address": cfg.Server.Listen})
	log.Error("Error", log15.Ctx{"code": s.ListenAndServe()})
	return 1
}

func defaultAction() int {
	fmt.Fprint(os.Stderr, "unsupported action; possible actions: server\n")
	return 1
}
