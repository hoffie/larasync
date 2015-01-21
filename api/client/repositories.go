package client

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/hoffie/larasync/api"
	. "github.com/hoffie/larasync/api/common"
)

// registerRequest builds a request for registering a new repository
func (c *Client) registerRequest(pubKey [PublicKeySize]byte) (*http.Request, error) {
	if len(c.adminSecret) == 0 {
		return nil, ErrMissingAdminSecret
	}
	body, err := json.Marshal(api.JSONRepository{
		PubKey: pubKey[:],
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PUT", c.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	SignWithPassphrase(req, c.adminSecret)
	return req, nil
}

// Register registers the current repository name with the server for the
// first time.
func (c *Client) Register(pubKey [PublicKeySize]byte) error {
	req, err := c.registerRequest(pubKey)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, http.StatusCreated)
	if err != nil {
		return err
	}
	return nil
}
