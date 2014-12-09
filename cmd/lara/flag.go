package main

import (
	"flag"
	"fmt"
)

var (
	configPath string
)

// makeFlagSet creates the dispatcher's flagset
func (d *Dispatcher) makeFlagSet(args []string) {
	name := "lara"
	if len(args) >= 1 {
		name = fmt.Sprintf("lara %s", args[0])
	}
	d.flags = flag.NewFlagSet(name, flag.ExitOnError)
	d.registerFlags()
	d.flags.Parse(args[1:])
}

// registerFlags is responsible for flag registration
func (d *Dispatcher) registerFlags() {
	d.flags.StringVar(&configPath, "config", "", "config file location")
}
