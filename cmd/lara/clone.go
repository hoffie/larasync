package main

import (
	"fmt"
	"os"

	apiclient "github.com/hoffie/larasync/api/client"
)

// syncAction implements the "lara clone" command.
func (d *Dispatcher) cloneAction() int {
	args := d.context.Args()
	if len(args) < 2 {
		fmt.Fprintln(d.stderr, "Error: Invalid Syntax")
		fmt.Fprintln(d.stderr, "Use: URL LOCAL-DIRECTORY")
		return 1
	}

	urlString := args[0]
	repoName := args[1]
	client, repo, err := apiclient.ImportAuthorization(repoName, urlString)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Unable to import authorization (%s)\n", err)
		return 1
	}
	dl := client.Downloader(repo)
	err = dl.GetAll()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Could not load data from server. (%s)\n", err)
		return 1
	}

	err = os.Chdir(repo.Path)
	if err != nil {
		fmt.Fprintf(d.stderr,
			"Error: Cannot chdir to repository root (%s)\n", err)
		return 1
	}
	return d.checkoutAllPathsAction()
}
