package main

import (
	"fmt"
	"os"

	"github.com/hoffie/larasync/repository"
)

// initAction initializes a new repository.
func (d *Dispatcher) initAction() int {
	numArgs := len(d.flags.Args())
	var target string
	if numArgs < 1 {
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Fprint(d.stderr, "Unable to get current directory\n")
			return 1
		}
		target = pwd
	} else {
		target = d.flags.Arg(0)
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
	err = repo.CreateEncryptionKey()
	if err != nil {
		fmt.Fprintf(d.stderr, "Unable to generate encryption key\n")
		return 1
	}
	err = repo.CreateSigningKey()
	if err != nil {
		fmt.Fprintf(d.stderr, "Unable to generate signing key\n")
		return 1
	}
	err = repo.CreateHashingKey()
	if err != nil {
		fmt.Fprintf(d.stderr, "Unable to generate hashing key\n")
		return 1
	}
	return 0
}
