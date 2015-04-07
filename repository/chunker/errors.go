package chunker

import (
	"errors"
)

var (
	// ErrBadChunkSize will be thrown if a too little chunk size is requested.
	// This is used by the Chunker implementation.
	ErrBadChunkSize = errors.New("bad chunk size (must be >16 bytes)")
)
