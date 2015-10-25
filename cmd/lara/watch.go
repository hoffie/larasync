package main

import (
	"fmt"

	"github.com/hoffie/larasync/repository"
	"github.com/hoffie/larasync/repository/watcher"
)

// watchCancel is used to cancel the watcher loop from outside.
// Primarily used for testing purposes.
var watchCancelChannel chan struct{}
// runningWatcher is the reference to the currently active watcher
// it is set for testing purposes.
var runningWatcher *watcher.Watcher = nil

// watchAction implements the "lara watch" command.
func (d *Dispatcher) watchAction() int {
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
	watcher, err := r.Watch()
	runningWatcher = watcher
	if err != nil {
		fmt.Fprint(d.stderr, "Error: Watching for file changes failed\n")
	}

	watchCancelChannel = make(chan struct{})
	for {
		exit := false
		select {
		case err = <-watcher.Errors:
			fmt.Fprint(d.stderr, "Error while watching for file changes:\n")
			fmt.Fprint(d.stderr, err)
		case <-watcher.Close:
			exit = true
			runningWatcher = nil
			break
		case <-watchCancelChannel:
			watcher.Stop()
		}
		if exit {
			break
		}
	}
	return 0
}
