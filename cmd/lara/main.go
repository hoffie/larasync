package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/inconshreveable/log15"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/repository"
)

var log = log15.New("module", "main")

// main is our service dispatcher.
func main() {
	dispatcher := &Dispatcher{stdin: os.Stdin, stdout: os.Stdout, stderr: os.Stderr}
	os.Exit(dispatcher.run(os.Args))
}

// Dispatcher is the environment for our command dispatcher and keeps
// references to the relevant external interfaces.
type Dispatcher struct {
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
	context  *cli.Context
	app      *cli.App
	exitCode int
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

	d.app = app
}

// run starts the cli with the entered arguments.
// returns the exit code.
func (d *Dispatcher) run(args []string) int {
	passArgs := []string{}
	progName := os.Args[0]

	if (len(args) > 0 && args[0] != progName) {
		passArgs = append(passArgs, progName)
	}
	passArgs = append(passArgs, args...)

	d.initApp()
	d.app.Run(passArgs)
	return d.exitCode
}

// setupLogging configures our loggers and sets up our subpackages to use
// it as well.
func (d *Dispatcher) setupLogging() {
	handler := log15.StreamHandler(os.Stderr, log15.LogfmtFormat())
	log.SetHandler(handler)
	repository.Log.SetHandler(handler)
	api.Log.SetHandler(handler)
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

// prompt outputs the given prompt text and waits for a value to be entered
// on the input stream.
func (d *Dispatcher) prompt(prompt string) ([]byte, error) {
	d.stdout.Write([]byte(prompt))
	switch d.stdin.(type) {
	case *os.File:
		return d.promptGetpass()
	}
	return d.promptUnsafe()
}

// promptGetpass reads a password from our input,
// attempting to hide the input if possible.
func (d *Dispatcher) promptGetpass() ([]byte, error) {
	file := d.stdin.(*os.File)
	fd := int(file.Fd())
	if !terminal.IsTerminal(fd) {
		return d.promptUnsafe()
	}
	defer d.stdout.Write([]byte("\n"))
	return terminal.ReadPassword(fd)
}

// promptUnsafe reads a password from our input in the standard way.
// It cannot hide the input; it's our fallback if no terminal
// is attached to the input stream.
func (d *Dispatcher) promptUnsafe() ([]byte, error) {
	r := bufio.NewReader(d.stdin)
	line, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return line[:len(line)-1], nil
}
