package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/inconshreveable/log15"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/repository"
)

var log = log15.New("module", "main")

// main is our service dispatcher.
func main() {
	d := &Dispatcher{stdin: os.Stdin, stdout: os.Stdout, stderr: os.Stderr}
	os.Exit(d.run(os.Args[1:]))
}

// Dispatcher is the environment for our command dispatcher and keeps
// references to the relevant external interfaces.
type Dispatcher struct {
	stdin  io.Reader
	stdout io.Writer
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
	case "admin-secret":
		cmd = d.adminSecretAction
	case "checkout":
		cmd = d.checkoutAction
	case "help":
		cmd = d.helpAction
	case "init":
		cmd = d.initAction
	case "register":
		cmd = d.registerAction
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
	fmt.Fprint(d.stderr, "  add          adds the current state of the given file or directory\n")
	fmt.Fprint(d.stderr, "  admin-secret asks for an admin secret outputs its hash\n")
	fmt.Fprint(d.stderr, "  checkout     (over)writes the given path with the repository's state\n")
	fmt.Fprint(d.stderr, "  help         this information\n")
	fmt.Fprint(d.stderr, "  init         initialize a new repository\n")
	fmt.Fprint(d.stderr, "  register     register this repository with a server\n")
	fmt.Fprint(d.stderr, "  server       run in server mode\n")
	return 0
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
