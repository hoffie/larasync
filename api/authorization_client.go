package api

import (
	"encoding/hex"
	"io"
	"net/http"
)

// putAuthorizationRequest generates a request to add new authorization
// data to the server
func (c *Client) putAuthorizationRequest(
	pubKey *[PublicKeySize]byte,
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
	pubKey *[PublicKeySize]byte,
	authorizationReader io.Reader,
) error {
	req, err := c.putAuthorizationRequest(pubKey, authorizationReader)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req, http.StatusOK, http.StatusCreated)
	return err
}

// getAuthorizationRequest generates a request to request a authorization
// from the server.
func (c *Client) getAuthorizationRequest(authorizationURL string,
	authPrivKey [PrivateKeySize]byte,) (*http.Request, error) {
	req, err := http.NewRequest(
		"GET",
		authorizationURL,
		nil,
	)
	if err != nil {
		return nil, err
	}

	SignWithKey(req, authPrivKey)
	return req, nil
}

func (c *Client) GetAuthorization(authorizationURL string,
	authPrivKey [PrivateKeySize]byte,) (io.Reader, error) {
	req, err := c.getAuthorizationRequest(authorizationURL, authPrivKey)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
