package config

import (
	"errors"
)

var (
	// ErrAdminPubkeyMissing is returned if no admin pubkey is specified.
	// It is used by the ServerConfig handling.
	ErrAdminPubkeyMissing = errors.New("empty admin pubkey")

	// ErrInvalidAdminPubkey is returned if decoding the admin pubkey fails.
	// It is used by the ServerConfig handling.
	ErrInvalidAdminPubkey = errors.New("invalid admin pubkey")

	// ErrTruncatedAdminPubkey is returned if the given admin pubkey is too short.
	// It is used by the ServerConfig handling.
	ErrTruncatedAdminPubkey = errors.New("admin pubkey too short")

	// ErrMissingBasePath is returned if no base path is configured.
	// It is used by the ServerConfig handling.
	ErrMissingBasePath = errors.New("missing basepath")

	// ErrBadBasePath is returned if the configured base path is not accessible.
	// It is used by the ServerConfig handling.
	ErrBadBasePath = errors.New("unaccessible basepath")
)
