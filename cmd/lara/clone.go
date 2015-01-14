package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/helpers/crypto"
	"github.com/hoffie/larasync/repository"
)

// syncAction implements the "lara clone" command.
func (d *Dispatcher) cloneAction() int {
	if len(d.flags.Args()) < 2 {
		fmt.Fprintln(d.stderr, "Error: Parameters invalid")
		fmt.Fprintln(d.stderr, "You have to pass the repository name to clone to as a first ")
		fmt.Fprintln(d.stderr, "argument and the authorization url as second argument.")
		return 1
	}

	args := d.flags.Args()

	repo := repository.NewClient(args[0])
	err := repo.Create()
	if err != nil && !os.IsExist(err) {
		fmt.Fprintln(d.stderr, "Error: Could not create repository: %s", err)
		return 1
	}

	urlString := args[1]
	u, err := url.Parse(urlString)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Could not parse url. (%s)\n", err)
		return 1
	}

	sc, err := repo.StateConfig()
	if err != nil {
		fmt.Fprintf(d.stderr, "unable to load state config (%s)\n", err)
		return 1
	}
	sc.DefaultServer = "http://" + u.Host + path.Dir(path.Dir(u.Path))
	err = sc.Save()
	if err != nil {
		fmt.Fprintf(d.stderr, "unable to save state config (%s)\n", err)
		return 1
	}

	client := api.NewClient(sc.DefaultServer)

	authURL, err := parseAuthURL(u)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Could not extract authorization information. (%s)\n", err)
		return 1
	}

	reader, err := client.GetAuthorization(authURL.URL.String(), authURL.SignKey)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Communication with server failed. (%s)\n", err)
		return 1
	}

	enc, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Could not get data from server. (%s)\n", err)
		return 1
	}

	box := crypto.NewBox(authURL.EncKey)
	data, err := box.DecryptContent(enc)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Decryption of response failed. (%s)\n", err)
		return 1
	}

	auth := &repository.Authorization{}
	_, err = auth.ReadFrom(bytes.NewBuffer(data))
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Authorization data could not be read. (%s)\n", err)
		return 1
	}

	err = repo.SetKeysFromAuth(auth)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Failed to store key data for the repository. (%s)\n", err)
		return 1
	}

	privKey, err := repo.GetSigningPrivateKey()
	if err != nil {
		fmt.Fprintf(d.stderr, "unable to get signing private key (%s)\n", err)
		return 1
	}
	client.SetSigningPrivateKey(privKey)

	dl := &downloader{client: client, r: repo}
	err = dl.getAll()
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
