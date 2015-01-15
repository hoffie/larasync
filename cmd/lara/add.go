package main

import (
	"fmt"

	"github.com/hoffie/larasync/repository"
)

// addAction adds the current state of the given file or directory to the repository.
func (d *Dispatcher) addAction() int {
	absPath, root, err := d.parseFirstPathArg()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: %s\n", err)
		return 1
	}
	r := repository.NewClient(root)
	err = r.AddItem(absPath)
	if err != nil {
		fmt.Fprintf(d.stderr, "Unable to add the given item to the repository (%s)\n", err)
		return 1
	}
	return 0
}
