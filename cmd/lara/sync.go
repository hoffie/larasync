package main

import (
	"fmt"

	"github.com/hoffie/larasync/repository"
)

// syncAction implements the "lara sync" command.
func (d *Dispatcher) syncAction() int {
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
	err = r.AddItem(root)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: adding local changes failed (%s)\n", err)
		return 1
	}
	dl := client.Downloader(r)
	ul := client.Uploader(r)
	if d.context.Bool("full") {
		err = dl.GetAll()
	} else {
		err = dl.GetDelta()
	}
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: syncing data from server failed (%s)\n", err)
		return 1
	}

	if d.context.Bool("full") {
		err = ul.PushAll()
	} else {
		err = ul.PushDelta()
	}
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: uploading data to the server failed (%s)\n", err)
		return 1
	}

	return d.checkoutAllPathsAction()
}
