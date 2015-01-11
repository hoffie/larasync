package main

import (
	"fmt"
	"os"

	"github.com/hoffie/larasync/repository"
)

// checkoutAction handles all "lara checkout" commands and dispatches to the
// appropriate sub-handlers.
func (d *Dispatcher) checkoutAction() int {
	numArgs := len(d.flags.Args())
	if numArgs > 1 {
		fmt.Fprintf(d.stderr, "Error: only one path is supported")
		return 1
	}
	if numArgs == 1 {
		return d.checkoutPathAction()
	}
	return d.checkoutAllPathsAction()
}

// checkoutPathAction handles "lara checkout path/to/file.txt"
func (d *Dispatcher) checkoutPathAction() int {
	absPath, root, err := d.parseFirstPathArg()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: %s\n", err)
		return 1
	}
	r := repository.New(root)
	err = r.CheckoutPath(absPath)
	if err != nil {
		fmt.Fprintf(d.stderr,
			"Unable to checkout the given path from the repository (%s)\n", err)
		return 1
	}
	return 0
}

// checkoutAllPathsAction handles "lara checkout" without any arguments.
func (d *Dispatcher) checkoutAllPathsAction() int {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: unable to get current working directory")
		return 1
	}
	root, err := repository.GetRoot(wd)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: unable to find a repository here")
		return 1
	}

	r := repository.New(root)
	err = r.CheckoutAllPaths()
	if err != nil {
		fmt.Fprintf(d.stderr,
			"Unable to checkout the given path from the repository (%s)\n", err)
		return 1
	}
	return 0
}
