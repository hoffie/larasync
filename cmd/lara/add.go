package main

import (
	"fmt"
	"path/filepath"

	"github.com/hoffie/larasync/repository"
)

// addAction adds the current state of the given file or directory to the repository.
func (d *Dispatcher) addAction() int {
	numArgs := len(d.flags.Args())
	if numArgs < 1 {
		fmt.Fprint(d.stderr, "No items specified\n")
		return 1
	}
	absPath, err := filepath.Abs(d.flags.Arg(0))
	if err != nil {
		fmt.Fprint(d.stderr, "Unable to resolve path\n")
		return 1
	}
	root, err := repository.GetRoot(absPath)
	if err != nil {
		fmt.Fprint(d.stderr, "Unable to find the repository root\n")
		return 1
	}
	r := repository.New(root)
	err = r.AddItem(absPath)
	if err != nil {
		fmt.Fprint(d.stderr, "Unable to add the given item to the repository\n")
		return 1
	}
	return 0
}
