package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	flags *flag.FlagSet
	configPath string
)

func makeFlagSet(args []string) *flag.FlagSet {
	name := ""
	if len(args) >= 2 {
		name = fmt.Sprintf("%s %s", args[0], args[1])
	} else if len(args) >= 1 {
		name = args[0]
	}
	return flag.NewFlagSet(name, flag.ExitOnError)
}

func init() {
	flags = makeFlagSet(os.Args)
	flags.StringVar(&configPath, "config", "", "config file location")
}
