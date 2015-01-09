package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/inconshreveable/log15"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/repository"
)

var log = log15.New("module", "main")

// main is our service dispatcher.
func main() {
	d := &Dispatcher{stderr: os.Stderr}
	os.Exit(d.run(os.Args[1:]))
}

// Dispatcher is the environment for our command dispatcher and keeps
// references to the relevant external interfaces.
type Dispatcher struct {
	stderr io.Writer
	flags  *flag.FlagSet
}

// run starts dispatching with the given args.
func (d *Dispatcher) run(args []string) int {
	d.makeFlagSet(args)
	if len(args) < 1 {
		fmt.Fprint(d.stderr, "Error: no action given\n")
		fmt.Fprint(d.stderr, "Please specify an action, e.g.\n\tlara help\n")
		return 1
	}
	action := args[0]
	cmd := d.defaultAction
	switch action {
	case "add":
		cmd = d.addAction
	case "checkout":
		cmd = d.checkoutAction
	case "help":
		cmd = d.helpAction
	case "init":
		cmd = d.initAction
	case "server":
		cmd = d.serverAction
	}
	return cmd()
}

// setupLogging configures our loggers and sets up our subpackages to use
// it as well.
func (d *Dispatcher) setupLogging() {
	handler := log15.StreamHandler(os.Stderr, log15.LogfmtFormat())
	log.SetHandler(handler)
	repository.Log.SetHandler(handler)
	api.Log.SetHandler(handler)
}

// helpAction outputs usage information.
func (d *Dispatcher) helpAction() int {
	fmt.Fprint(d.stderr, "Syntax: lara ACTION\n\n")
	fmt.Fprint(d.stderr, "Possible actions:\n")
	fmt.Fprint(d.stderr, "\tadd      adds the current state of the given file or directory\n")
	fmt.Fprint(d.stderr, "\tcheckout (over)writes the given path with the repository's state\n")
	fmt.Fprint(d.stderr, "\thelp     this information\n")
	fmt.Fprint(d.stderr, "\tinit     initialize a new repository\n")
	fmt.Fprint(d.stderr, "\tserver   run in server mode\n")
	return 0
}

// serverAction starts the server process.
func (d *Dispatcher) serverAction() int {
	d.setupLogging()
	cfg, err := getServerConfig()
	if err != nil {
		log.Error("unable to parse configuration", log15.Ctx{"error": err})
		return 1
	}
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

// defaultAction is invoked for all unknown actions.
func (d *Dispatcher) defaultAction() int {
	fmt.Fprint(d.stderr, "Error: unknown action\n")
	fmt.Fprint(d.stderr, "Please specify a valid action, see \n\tlara help\n")
	return 1
}

// parseFirstPathArg takes the first command line argument and returns its
// absolute value along with the associated repository root.
func (d *Dispatcher) parseFirstPathArg() (string, string, error) {
	numArgs := len(d.flags.Args())
	if numArgs < 1 {
		return "", "", errors.New("no path specified")
	}
	absPath, err := filepath.Abs(d.flags.Arg(0))
	if err != nil {
		return "", "", errors.New("unable to resolve path")
	}
	root, err := repository.GetRoot(absPath)
	if err != nil {
		return "", "", errors.New("unable to find the repository root")
	}
	return absPath, root, nil
}
