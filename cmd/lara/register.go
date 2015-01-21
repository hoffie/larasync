package main

import (
	"fmt"

	"github.com/hoffie/larasync/api/client"
	"github.com/hoffie/larasync/repository"
)

// registerAction implements "lara register HOST NAME"
func (d *Dispatcher) registerAction() int {
	args := d.context.Args()
	if len(args) != 2 {
		fmt.Fprint(d.stderr,
			"Error: please specify the remote host and a name\n"+
				"\te.g. lara register example.org:14124 foo\n")
		return 1
	}
	root, err := d.getRootFromWd()
	if err != nil {
		return 1
	}
	netloc := args[0]
	repoName := args[1]
	adminSecret, err := d.promptPassword("Admin secret: ")
	if err != nil {
		fmt.Fprint(d.stderr, "Error: unable to read the admin secret\n")
		return 1
	}
	r := repository.NewClient(root)
	pubKey, err := r.GetSigningPublicKey()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: unable to retrieve local signing public key\n")
		return 1
	}

	url := client.NetlocToURL(netloc, repoName)

	sc, err := r.StateConfig()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: unable to load repo state (%s)\n", err)
		return 1
	}
	defaultServer := sc.DefaultServer
	defaultServer.URL = url

	client := d.clientForState(sc)
	client.SetAdminSecret(adminSecret)
	err = client.Register(pubKey)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: unable to register (%s)\n", err)
		return 1
	}
	err = sc.Save()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: unable to save repo state (%s)\n", err)
		return 1
	}
	fmt.Fprintf(d.stdout, "Successfully registered\n")
	return 0
}
