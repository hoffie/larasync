package main

import (
	"fmt"

	"github.com/hoffie/larasync/repository"
)

// pushAction implements "lara push"
func (d *Dispatcher) pushAction() int {
	if len(d.context.Args()) != 0 {
		fmt.Fprint(d.stderr, "Error: this command takes no arguments\n")
		return 1
	}
	root, err := d.getRootFromWd()
	if err != nil {
		return 1
	}
	r := repository.NewClient(root)

	client, err := clientFor(r)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: %s\n",
			err)
		return 1
	}

	ul := &uploader{
		client: client,
		r:      r,
	}
	err = ul.pushAll()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: %s\n",
			err)
		return 1
	}

	return 0
}
