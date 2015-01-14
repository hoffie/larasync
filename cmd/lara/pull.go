package main

import (
	"fmt"

	"github.com/hoffie/larasync/repository"
)

// pullAction implements "lara pull"
func (d *Dispatcher) pullAction() int {
	if len(d.flags.Args()) != 0 {
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
		fmt.Fprint(d.stderr, err)
		return 1
	}
	dl := downloader{client: client, r: r}
	err = dl.getAll()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: pull failed (%s)\n", err)
		return 1
	}
	return 0
}
