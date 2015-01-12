package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

var (
	// ErrMissingAdminSecret is returned if a method requiring the admin
	// secret (such as Register()) is called without having set one first.
	ErrMissingAdminSecret = errors.New("missing adminSecret")

	// ErrUnexpectedStatus is returned whenever the request did not yield
	// the expected HTTP status code.
	ErrUnexpectedStatus = errors.New("unexpected http status")
)

// Client provides convenience methods for accessing an api.Server
// over HTTP.
type Client struct {
	http        *http.Client
	baseURL     string
	netloc      string
	adminSecret []byte
}

// NewClient returns a new Client instance.
func NewClient(netloc, repoName string) *Client {
	return &Client{
		http: &http.Client{},
		// IMPROVEMENT: use mux router to generate URLs
		netloc:  netloc,
		baseURL: "http://" + netloc + "/repositories/" + repoName,
	}
}

// SetAdminSecret sets the admin secret to use (e.g. for Register()).
func (c *Client) SetAdminSecret(s []byte) {
	c.adminSecret = s
}

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
	req, err := http.NewRequest("PUT", c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", c.netloc)
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
