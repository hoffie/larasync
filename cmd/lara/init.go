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
	repo := repository.New(target)
	err := repo.CreateManagementDir()
	if err != nil {
		fmt.Fprint(d.stderr, "Unable to create management directory\n")
		return 1
	}
	err = repo.CreateKeys()
	if err != nil {
		fmt.Fprintf(d.stderr, "Unable to generate encryption keys\n")
		return 1
	}
	return 0
}
