package main

import (
	"fmt"

	"github.com/hoffie/larasync/api"
)

// adminSecretAction implements "lara admin-secret"
func (d *Dispatcher) adminSecretAction() int {
	if len(d.context.Args()) != 0 {
		fmt.Fprint(d.stderr, "Error: this command takes no args\n")
		return 1
	}
	adminSecret, err := d.promptPassword("Admin secret: ")
	if err != nil {
		fmt.Fprint(d.stderr, "Error: unable to read the admin secret\n")
		return 1
	}
	adminPubkey, err := api.GetAdminSecretPubkey(adminSecret)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: public key retrieval failed(%s)\n", err)
		return 1
	}
	fmt.Fprintf(d.stdout, "# Enter the following value into your server config\n%x\n",
		adminPubkey)
	return 0
}
