package main

import (
	"bytes"
	"crypto/rand"
	"fmt"

	apiclient "github.com/hoffie/larasync/api/client"
	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
	"github.com/hoffie/larasync/repository"
)

// authorizeNewClient is the command line handler for a specific
// repository to put a signed authorization signature to the server.
func (d *Dispatcher) authorizeNewClientAction() int {
	root, err := d.getRootFromWd()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: %s\n", err)
		return 1
	}

	var encryptionKey [EncryptionKeySize]byte
	_, err = rand.Read(encryptionKey[:])
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Encryption key generating error: %s\n", err)
		return 1
	}

	signingPubKey, signingPrivKey, err := edhelpers.GenerateKey()

	if err != nil || signingPubKey == nil || signingPrivKey == nil {
		fmt.Fprintf(d.stderr, "Error: Signature key generation error: %s\n", err)
		return 1
	}
	r := repository.NewClient(root)
	auth, err := r.NewAuthorization()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Could not create authorization package: %s\n", err)
		return 1
	}

	authorizationBytes, err := r.SerializeAuthorization(encryptionKey, auth)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Could not encrypt authorization information: %s\n", err)
		return 1
	}

	client, err := d.clientFor(r)

	if err != nil {
		fmt.Fprintf(d.stderr, "Error: %s\n", err)
		return 1
	}

	defaultServer := d.sc.DefaultServer
	authURL, err := apiclient.NewAuthURL(client.BaseURL, signingPrivKey, &encryptionKey,
		defaultServer.Fingerprint)
	if err != nil {
		fmt.Fprintln(d.stderr, "Error: authorization url could not be generated.")
		return 1
	}

	err = client.PutAuthorization(signingPubKey, bytes.NewBuffer(authorizationBytes))
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Server communication failed (%s)\n", err)
		return 1
	}

	fmt.Fprintln(d.stdout, "New authorization request completed")
	fmt.Fprintln(d.stdout, authURL.String())
	return 0
}
