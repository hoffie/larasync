package main

import (
	"github.com/codegangsta/cli"
)

// globalFlags returns the flags that should be
// registered to the main config file.
func (d *Dispatcher) globalFlags() []cli.Flag {
	return []cli.Flag{}
}

// serverFlags returns the flags that should be
// registered as flags available in the "server"
// subcommand.
func (d *Dispatcher) serverFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "",
			Usage: "config file location",
		},
	}
}

// pushFlags returns the flags that should be
// registered as flags available in the "push"
// subcommand.
func (d *Dispatcher) pushFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "full",
			Usage: "forces a full synchronization to the server",
		},
	}
}
