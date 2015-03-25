package repository

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrInvalidPublicKeySize will get thrown if a string is passed
	// which couldn't be encoded to the correct size to pass it as a
	// Public Key signature.
	ErrInvalidPublicKeySize = errors.New("Invalid public key size.")
	// ErrSignatureVerification gets returned if a signature of a signed NIB could
	// not be verified.
	ErrSignatureVerification = errors.New("Signature verification failed")
	// ErrUnMarshalling gets returned if a NIB could not be extracted from stored
	// bytes.
	ErrUnMarshalling = errors.New("Couldn't extract item from byte stream")
	// ErrTransactionNotExists is thrown if a transaction could not be found. This is used
	// by the transaction manager.
	ErrTransactionNotExists = errors.New("Transaction does not exist in repository.")
	// ErrNIBConflict is returned when attempting to import a NIB
	// which may not be fast-forwarded.
	ErrNIBConflict = errors.New("NIB conflict (cannot fast forward)")
	// ErrRefusingWorkOnDotLara is thrown when an attempt is made to add the
	// management directory or some content to the repository.
	ErrRefusingWorkOnDotLara = errors.New("will not work on .lara")
	// ErrWorkDirConflict is being returned if a checkout path has changed data.
	ErrWorkDirConflict = errors.New("workdir conflict")
)

// NewErrNIBContentMissing returns a new ErrNIBContentMissing Error with the passed
// content IDs marked as missing.
func NewErrNIBContentMissing(contentIDs []string) *ErrNIBContentMissing {
	return &ErrNIBContentMissing{
		contentIDs: contentIDs,
	}
}

// ErrNIBContentMissing is returned when trying to add a NIB to the repository and
// content IDs are missing.
type ErrNIBContentMissing struct {
	contentIDs []string
}

// Error returns the error message which encodes the not found content ID.
func (e *ErrNIBContentMissing) Error() string {
	return fmt.Sprintf(
		"Content of passed NIB is not stored yet. Missing contentIDs: %s",
		strings.Join(e.contentIDs, ", "),
	)
}

// MissingContentIDs returns all ids which couldn't be resolved in the
// repository.
func (e *ErrNIBContentMissing) MissingContentIDs() []string {
	return e.contentIDs
}

// IsNIBContentMissing checks if the passed error is a nibContentMissing error.
func IsNIBContentMissing(err error) bool {
	switch err.(type) {
	case nil:
		return false
	case *ErrNIBContentMissing:
		return true
	default:
		return false
	}
}
