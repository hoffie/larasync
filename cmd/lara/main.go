package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/inconshreveable/log15"

	"github.com/hoffie/larasync/api/client"
	"github.com/hoffie/larasync/api/server"
	"github.com/hoffie/larasync/repository"
)

var log = log15.New("module", "main")

// main is our service dispatcher.
func main() {
	dispatcher := &Dispatcher{stdin: os.Stdin, stdout: os.Stdout, stderr: os.Stderr}
	args := []string{}
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}
	os.Exit(dispatcher.run(args))
}

// Dispatcher is the environment for our command dispatcher and keeps
// references to the relevant external interfaces.
type Dispatcher struct {
	stdin         io.Reader
	stdout        io.Writer
	stderr        io.Writer
	context       *cli.Context
	app           *cli.App
	sc            *repository.StateConfig
	serverCfgPath string
	exitCode      int
}

// initApp initializes the app structure.
func (d *Dispatcher) initApp() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Usage = "least authority rest assured synchronization"
	app.Version = "pre-build"
	app.Author = "The larasync team"
	app.Email = "team@larasync.org"
	app.Commands = d.cmdActions()
	app.Flags = d.globalFlags()
	app.Writer = d.stdout
	app.CommandNotFound = func(ctx *cli.Context, cmd string) {
		fmt.Fprint(d.stderr, "No such command; try help\n")
		d.exitCode = 1
	}

	d.app = app
}

// run starts the cli with the entered arguments.
// returns the exit code.
func (d *Dispatcher) run(args []string) int {
	passArgs := []string{"lara"}

	passArgs = append(passArgs, args...)

	d.initApp()
	if len(args) == 0 {
		d.exitCode = 1
	}
	err := d.app.Run(passArgs)
	if err != nil {
		d.exitCode = 1
	}
	return d.exitCode
}

// setupLogging configures our loggers and sets up our subpackages to use
// it as well.
func (d *Dispatcher) setupLogging() {
	handler := log15.StreamHandler(d.stderr, log15.LogfmtFormat())
	log.SetHandler(handler)
	repository.Log.SetHandler(handler)
	server.Log.SetHandler(handler)
	client.Log.SetHandler(handler)
}

// parseFirstPathArg takes the first command line argument and returns its
// absolute value along with the associated repository root.
func (d *Dispatcher) parseFirstPathArg() (string, string, error) {
	args := d.context.Args()
	numArgs := len(args)
	if numArgs < 1 {
		return "", "", errors.New("no path specified")
	}

	absPath, err := filepath.Abs(args[0])
	if err != nil {
		return "", "", errors.New("unable to resolve path")
	}
	root, err := repository.GetRoot(absPath)
	if err != nil {
		return "", "", errors.New("unable to find the repository root")
	}
	return absPath, root, nil
}

// getRootFromWd tries to find a repository root starting in the current
// working directory.
// Errors out, if none can be found.
func (d *Dispatcher) getRootFromWd() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: unable to get current working directory")
		return "", errors.New("unable to get cwd")
	}
	root, err := repository.GetRoot(wd)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: unable to find a repository here")
		return "", errors.New("unable to find a repository here")
	}
	return root, nil
}
