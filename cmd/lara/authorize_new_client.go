package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/hoffie/larasync/api"
	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
	"github.com/hoffie/larasync/repository"
)

func authorizationURLFor(c *api.Client, signingPrivKey *[PrivateKeySize]byte, encryptionKey *[EncryptionKeySize]byte) string {
	signingPrivKeyString := hex.EncodeToString(signingPrivKey[:])
	encryptionKeyString := hex.EncodeToString(encryptionKey[:])
	pubKey := edhelpers.GetPublicKeyFromPrivate(*signingPrivKey)
	pubKeyString := hex.EncodeToString(pubKey[:])
	return fmt.Sprintf("%s/authorizations/%s#AuthEncKey=%s&AuthSignKey=%s",
		c.BaseURL, pubKeyString, encryptionKeyString, signingPrivKeyString)
}

// authorizeNewClient is the command line handler for a specific
// repository to put a signed authorization signature to the server.
func (d *Dispatcher) authorizeNewClient() int {
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
		fmt.Fprintf(d.stderr, "Error: Signature key generating error: %s\n", err)
	}
	r := repository.New(root)
	auth, err := r.CurrentAuthorization()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Could not fetch current authorization from repository: %s\n", err)
		return 1
	}

	authorizationBytes, err := r.SerializeAuthorization(encryptionKey, auth)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Could not encrypt authorization information: %s\n", err)
		return 1
	}

	client, err := clientFor(r)

	if err != nil {
		fmt.Fprintf(d.stderr, "Error: %s\n", err)
		return 1
	}

	err = client.PutAuthorization(signingPubKey, bytes.NewBuffer(authorizationBytes))
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: Server communication failed (%s)\n", err)
		return 1
	}

	fmt.Fprintln(d.stdout, "New authorization request completed")
	fmt.Fprintln(d.stdout, authorizationURLFor(client, signingPrivKey, &encryptionKey))
	return 0
}
