package main

import (
	"fmt"

	"github.com/hoffie/larasync/repository"
)

// resetFingerprintAction implements "reset-fingerprint"
func (d *Dispatcher) resetFingerprintAction() int {
	root, err := d.getRootFromWd()
	if err != nil {
		return 1
	}
	fmt.Fprintf(d.stdout,
		"Clearing out the saved server fingerprint should only be necessary when you\n"+
			"know that your server's configuration has changed.\n")
	res, err := d.promptCleartext("Really reset fingerprint? ")
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: unable to read from terminal (%s)\n", err)
		return 1
	}
	if string(res) != "y" {
		return 0
	}
	r := repository.NewClient(root)
	sc, err := r.StateConfig()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Unable to load state config (%s)\n", err)
		return 1
	}
	sc.DefaultServer.Fingerprint = ""
	err = sc.Save()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Unable to save state config (%s)\n", err)
		return 1
	}
	return 0
}
