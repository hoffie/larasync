package api

import (
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
	BaseURL     string
	netloc      string
	adminSecret []byte
}

// NetlocToURL returns the URL matching the given netloc
func NetlocToURL(netloc, repoName string) string {
	// IMPROVEMENT: use mux router to generate URLs
	return "http://" + netloc + "/repositories/" + repoName
}

// NewClient returns a new Client instance.
func NewClient(url string) *Client {
	return &Client{
		http:    &http.Client{},
		BaseURL: url,
	}
}

// SetAdminSecret sets the admin secret to use (e.g. for Register()).
func (c *Client) SetAdminSecret(s []byte) {
	c.adminSecret = s
}


