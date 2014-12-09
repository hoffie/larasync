package main

import (
	"os"
	"fmt"
	"io"

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
		fmt.Fprint(os.Stderr, "Error: no action given\n")
		fmt.Fprint(os.Stderr, "Please specify an action, e.g.\n\tlara help\n")
		return 1
	}
	action := args[0]
	if len(args) > 1 {
		flags.Parse(args[1:])
	}
	cmd := defaultAction
	switch action {
	case "help":
		cmd = helpAction
	case "init":
		cmd = initAction
	case "server":
		cmd = serverAction
	}
	return cmd(os.Stderr)
}

func setupLogging() {
	handler := log15.StreamHandler(os.Stderr, log15.LogfmtFormat())
	log.SetHandler(handler)
	repository.Log.SetHandler(handler)
	api.Log.SetHandler(handler)
}

func helpAction(out io.Writer) int {
	fmt.Fprint(out, "Syntax: lara ACTION\n\n")
	fmt.Fprint(out, "Possible actions:\n")
	fmt.Fprint(out, "\thelp\tthis information\n")
	fmt.Fprint(out, "\tinit\tinitialize a new repository\n")
	fmt.Fprint(out, "\tserver\trun in server mode\n")
	return 0
}

func serverAction(out io.Writer) int {
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

func defaultAction(out io.Writer) int {
	fmt.Fprint(out, "Error: unknown action\n")
	fmt.Fprint(out, "Please specify a valid action, see \n\tlara help\n")
	return 1
}
