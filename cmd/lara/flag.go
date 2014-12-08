package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	flags = flag.NewFlagSet(fmt.Sprintf("%s %s", os.Args[0], os.Args[1]),
		flag.ExitOnError)
	configPath string
)

func init() {
	flags.StringVar(&configPath, "config", "", "config file location")
}
