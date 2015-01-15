package main

import (
	"github.com/codegangsta/cli"
)

var (
	configPath string
)

// globalFlags returns the flags that should be
// registered to the main config file.
func (d *Dispatcher) globalFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "",
			Usage: "config file location",
		},
	}
}
