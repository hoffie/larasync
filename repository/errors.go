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
	// ErrBadChunkSize will be thrown if a too little chunk size is requested.
	// This is used by the Chunker implementation.
	ErrBadChunkSize = errors.New("bad chunk size (must be >16 bytes)")
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
)

// NIBContentMissing is returned to
type NIBContentMissing struct {
	contentIDs []string
}

// Error returns the error message which encodes the not found content ID.
func (e *NIBContentMissing) Error() string {
	return fmt.Sprintf(
		"Content of passed NIB is not stored yet. Missing contentIDs: %s",
		strings.Join(e.contentIDs, ", "),
	)
}

// MissingContentIDs returns all ids which couldn't be resolved in the
// repository.
func (e *NIBContentMissing) MissingContentIDs() []string {
	return e.contentIDs
}

// IsNIBContentMissing checks if the passed error is a nibContentMissing error.
func IsNIBContentMissing(err error) bool {
	switch err.(type) {
	case nil:
		return false
	case *NIBContentMissing:
		return true
	default:
		return false
	}
}
