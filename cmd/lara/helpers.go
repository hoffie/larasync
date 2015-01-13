package lara

import (
	"fmt"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/repository"
)

// clientFor returns the Client which is configured to communicate
// with the given server repository.
func clientFor(r *repository.Repository) (*api.Client, error) {
	sc, err := r.StateConfig()
	if err != nil {
		return nil, fmt.Errorf("Error: unable to load state config (%s)", err)
	}
	if sc.DefaultServer == "" {
		return nil, fmt.Errorf("Error: no default server configured (state)")
	}
	privKey, err := r.GetSigningPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("Error: unable to get signing private key (%s)", err)
	}
	client := api.NewClient(sc.DefaultServer)
	client.SetSigningPrivateKey(privKey)

	return client, nil
}
