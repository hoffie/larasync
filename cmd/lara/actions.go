package main

import (
	"github.com/codegangsta/cli"
)

// cmdAction has to be implemented for commands executions.
type cmdAction func() int

// wrapAction is used to bootstrap the dispatcher before running the cmd
// function.
func (d *Dispatcher) wrapAction(action cmdAction) func(c *cli.Context) {
	return func(c *cli.Context) {
		d.context = c
		d.exitCode = action()
	}
}

// cmdActions returns the command line arguments available to the CLI.
func (d *Dispatcher) cmdActions() []cli.Command {
	return []cli.Command{
		{
			Name:   "add",
			Usage:  "adds the current state of the given file or directory.",
			Action: d.wrapAction(d.addAction),
		},
		{
			Name:   "admin-secret",
			Usage:  "asks for an admin secret outputs its hash.",
			Action: d.wrapAction(d.adminSecretAction),
		},
		{
			Name:   "authorize-new-client",
			Usage:  "initializes a new authorization variable for a new client.",
			Action: d.wrapAction(d.authorizeNewClientAction),
		},
		{
			Name:   "checkout",
			Usage:  "(over)writes the given path with the repository's state.",
			Action: d.wrapAction(d.checkoutAction),
		},
		{
			Name:   "clone",
			Usage:  "downloads an already initialized repository",
			Action: d.wrapAction(d.cloneAction),
		},
		{
			Name:   "init",
			Usage:  "initialize a new repository.",
			Action: d.wrapAction(d.initAction),
		},
		{
			Name:   "pull",
			Usage:  "downlodas the current state from the server.",
			Action: d.wrapAction(d.pullAction),
			Flags:  d.pullFlags(),
		},
		{
			Name:   "push",
			Usage:  "uploads the current state to the server.",
			Action: d.wrapAction(d.pushAction),
			Flags:  d.pushFlags(),
		},
		{
			Name:   "register",
			Usage:  "register this repository with a server.",
			Action: d.wrapAction(d.registerAction),
		},
		{
			Name:   "reset-fingerprint",
			Usage:  "resets the stored server fingerprint",
			Action: d.wrapAction(d.resetFingerprintAction),
		},
		{
			Name:   "server",
			Usage:  "run in server mode.",
			Action: d.wrapAction(d.serverAction),
			Flags:  d.serverFlags(),
		},
		{
			Name:   "server-fingerprint",
			Usage:  "print server certificate's public key fingerprint",
			Action: d.wrapAction(d.serverFingerprintAction),
		},
		{
			Name:   "sync",
			Usage:  "uploads and downloads all files from and to the repository.",
			Action: d.wrapAction(d.syncAction),
			Flags:  d.syncFlags(),
		},
	}
}
