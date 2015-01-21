package client

import (
	"errors"
)

var (
	// ErrMissingAdminSecret is returned if a method requiring the admin
	// secret (such as Register()) is called without having set one first.
	ErrMissingAdminSecret = errors.New("missing adminSecret")

	// ErrUnexpectedStatus is returned whenever the request did not yield
	// the expected HTTP status code.
	ErrUnexpectedStatus = errors.New("unexpected http status")
)
