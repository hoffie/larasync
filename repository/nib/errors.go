package nib

import (
	"errors"
)

var (
	// ErrNoRevision is returned if no such revision can be found. This is thrown
	// durint the NIB validation process.
	ErrNoRevision = errors.New("no revision")
)
