package repository

import (
	"errors"
	"io"
	"os"
)

// Chunker returns the contents of the given file path in chunkSize-sized
// chunks.
type Chunker struct {
	file      *os.File
	finished  bool
	chunkSize uint64
}

// ErrBadChunkSize will be thrown if a too little chunk size is requested
var ErrBadChunkSize = errors.New("bad chunk size (must be >16 bytes)")

// NewChunker returns a new chunker instance for the given file path
// and the given chunk size.
func NewChunker(path string, chunkSize uint64) (*Chunker, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	if chunkSize < 16 {
		return nil, ErrBadChunkSize
	}

	c := &Chunker{
		file:      file,
		finished:  false,
		chunkSize: chunkSize,
	}
	return c, nil
}

// HasNext returns true if an upcoming Next() call is expected
// to return more data.
func (c *Chunker) HasNext() bool {
	return !c.finished
}

// Next returns the next part of the file, with at most chunkSize bytes.
func (c *Chunker) Next() ([]byte, error) {
	buf := make([]byte, c.chunkSize)
	numBytes, err := c.file.Read(buf)
	if uint64(numBytes) < c.chunkSize || err != nil {
		c.finished = true
	}
	if err != nil || numBytes == 0 {
		c.file.Close()
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buf[:numBytes], nil
}

// Close cleans up the chunker after usage.
//
// Use the c := NewChunker(); defer c.Close() pattern
func (c *Chunker) Close() {
	c.file.Close()
}
