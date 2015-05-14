package main

import (
	"fmt"
	"os"

	"github.com/hoffie/larasync/repository"
)

// initAction initializes a new repository.
func (d *Dispatcher) initAction() int {
	args := d.context.Args()
	numArgs := len(args)
	var target string
	if numArgs < 1 {
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Fprint(d.stderr, "Unable to get current directory\n")
			return 1
		}
		target = pwd
	} else {
		target = args[0]
		err := os.Mkdir(target, 0700)
		if err != nil && os.IsExist(err) {
			fmt.Fprint(d.stderr, "Unable to create directory\n")
			return 1
		}
	}
	repo, err := repository.NewClient(target)
	if err != nil {
		fmt.Fprintf(d.stderr, "Unable to initialize repository\n")
		return 1
	}
	err = repo.Create()
	if err != nil {
		fmt.Fprint(d.stderr, "Unable to create management repository\n")
		return 1
	}
	return 0
}
