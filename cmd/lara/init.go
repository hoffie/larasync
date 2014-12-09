package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/hoffie/larasync/repository"
)

func initAction(out io.Writer, flags *flag.FlagSet) int {
	numArgs := len(flags.Args())
	var target string
	if numArgs < 1 {
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Fprint(out, "Unable to get current directory\n")
			return 1
		}
		target = pwd
	} else {
		target = flags.Arg(0)
		err := os.Mkdir(target, 0700)
		if err != nil && err != os.ErrExist {
			fmt.Fprint(out, "Unable to create directory\n")
			return 1
		}
	}
	repo := repository.New(target)
	err := repo.CreateManagementDir()
	if err != nil {
		fmt.Fprint(out, "Unable to create management directory\n")
		return 1
	}
	return 0
}
