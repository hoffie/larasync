package main

import (
	"fmt"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/repository"
)

// pushAction implements "lara push"
func (d *Dispatcher) pushAction() int {
	if len(d.flags.Args()) != 0 {
		fmt.Fprint(d.stderr, "Error: this command takes no arguments\n")
		return 1
	}
	root, err := d.getRootFromWd()
	if err != nil {
		return 1
	}
	r := repository.New(root)
	sc, err := r.StateConfig()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: unable to load state config (%s)\n", err)
		return 1
	}
	if sc.DefaultServer == "" {
		fmt.Fprintf(d.stderr, "Error: no default server configured (state)\n")
		return 1
	}
	client := api.NewClient(sc.DefaultServer)
	_ = client
	//FIXME:
	// - iterate over all nibs, upload them
	// - iterate over all objects, upload them
	return 0
}
