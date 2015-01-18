package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"

	"github.com/hoffie/larasync/helpers/crypto"
	"github.com/hoffie/larasync/repository"
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
	repo := repository.NewClient(repoName)
	err := repo.Create()
	if err != nil && !os.IsExist(err) {
		fmt.Fprintf(d.stderr, "Error: Could not create repository: %s\n", err)
		return 1
	}

	u, err := url.Parse(urlString)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Could not parse url. (%s)\n", err)
		return 1
	}
	authURL, err := parseAuthURL(u)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Could not extract authorization information. (%s)\n", err)
		return 1
	}

	sc, err := repo.StateConfig()
	if err != nil {
		fmt.Fprintf(d.stderr, "unable to load state config (%s)\n", err)
		return 1
	}
	sc.DefaultServer = "https://" + u.Host + path.Dir(path.Dir(u.Path))
	sc.DefaultServerFingerprint = authURL.Fingerprint
	err = sc.Save()
	if err != nil {
		fmt.Fprintf(d.stderr, "unable to save state config (%s)\n", err)
		return 1
	}

	client := d.clientForState(sc)

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
