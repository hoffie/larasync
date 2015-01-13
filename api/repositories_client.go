package api

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// registerRequest builds a request for registering a new repository
func (c *Client) registerRequest(pubKey [PublicKeySize]byte) (*http.Request, error) {
	if len(c.adminSecret) == 0 {
		return nil, ErrMissingAdminSecret
	}
	body, err := json.Marshal(JSONRepository{
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
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return ErrUnexpectedStatus
	}
	return nil
}
