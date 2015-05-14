package main

import (
	"fmt"

	"github.com/hoffie/larasync/repository"
)

// pullAction implements "lara pull"
func (d *Dispatcher) pullAction() int {
	if len(d.context.Args()) != 0 {
		fmt.Fprint(d.stderr, "Error: this command takes no arguments\n")
		return 1
	}
	root, err := d.getRootFromWd()
	if err != nil {
		return 1
	}
	r, err := repository.NewClient(root)
	if err != nil {
		fmt.Fprint(d.stderr, err)
		return 1
	}
	client, err := d.clientFor(r)
	if err != nil {
		fmt.Fprint(d.stderr, err)
		return 1
	}
	dl := client.Downloader(r)

	if d.context.Bool("full") {
		log.Info("Full download requested.")
		err = dl.GetAll()
	} else {
		log.Info("Delta download requested.")
		err = dl.GetDelta()
	}
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: pull failed (%s)\n", err)
		return 1
	}
	return 0
}
