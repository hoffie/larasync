package main

import (
	"fmt"

	"github.com/hoffie/larasync/repository"
)

//
func (d *Dispatcher) checkoutAction() int {
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
