package main

import (
	"fmt"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/repository"
)

const (
	// PrivateKeySize is the size of the key used for signing.
	PrivateKeySize = repository.PrivateKeySize
	// PublicKeySize ist the size of the key used to verify the signature.
	PublicKeySize = repository.PublicKeySize
	// EncryptionKeySize is the key size used for encryption purposes.
	EncryptionKeySize = repository.EncryptionKeySize
)

// clientFor returns the Client which is configured to communicate
// with the given server repository.
func clientFor(r *repository.ClientRepository) (*api.Client, error) {
	sc, err := r.StateConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to load state config (%s)", err)
	}
	if sc.DefaultServer == "" {
		return nil, fmt.Errorf("no default server configured (state)")
	}
	privKey, err := r.GetSigningPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("unable to get signing private key (%s)", err)
	}
	client := api.NewClient(sc.DefaultServer)
	client.SetSigningPrivateKey(privKey)

	return client, nil
}
