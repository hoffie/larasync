package main

import (
	"fmt"

	//"github.com/hoffie/larasync/repository"
)

// registerAction implements "lara register URL NAME"
func (d *Dispatcher) registerAction() int {
	if len(d.flags.Args()) != 2 {
		fmt.Fprint(d.stderr,
			"Error: please specify the remote URL and a name\n")
		return 1
	}
	return 0
}
