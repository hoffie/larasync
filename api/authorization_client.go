package api

import (
	"encoding/hex"
	"io"
	"net/http"
)

// putAuthorizationRequest generates a request to add new authorization
// data to the server
func (c *Client) putAuthorizationRequest(
	pubKey [PublicKeySize]byte,
	authorizationReader io.Reader,
) (*http.Request, error) {
	pubKeyString := hex.EncodeToString(pubKey[:])

	req, err := http.NewRequest(
		"PUT",
		c.BaseURL+"/authorizations/"+pubKeyString,
		authorizationReader,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	SignWithKey(req, c.signingPrivateKey)
	return req, nil
}

// PutAuthorization adds a new authorization assignment
// for the passed public key to the server.
func (c *Client) PutAuthorization(
	pubKey [PublicKeySize]byte,
	authorizationReader io.Reader,
) error {
	req, err := c.putAuthorizationRequest(pubKey, authorizationReader)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req, http.StatusOK, http.StatusCreated)
	return err
}
