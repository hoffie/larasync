package bincontainer

import (
	"encoding/binary"
	"errors"
	"io"
)

// A Decoder is able to reconstruct chunks of binary data from a stream which was
// previously encoded by an Encoder.
// NOTE: A maximum length of 2**32 (~4GB) is supported.
type Decoder struct {
	r io.Reader
}

// ErrIncomplete is thrown whenever the Reader is unable to supply further data,
// although the protocol assumes more.
var ErrIncomplete = errors.New("incomplete chunk")

// NewDecoder returns a new Decoder instance.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// ReadChunk attempts to read the next chunk of data from the underlying reader.
// It either returns the full chunk or an error.
// If the underlying writer is closed after reading a complete chunk, EOF is returned.
// In all other cases, another error (such as ErrIncomplete) is returned.
func (e *Decoder) ReadChunk() ([]byte, error) {
	length, err := e.readLength()
	if err != nil {
		return nil, err
	}
	return e.readData(length)
}

// readData attempts to read the requested number of bytes from the reader and
// returns it.
func (e *Decoder) readData(length uint32) ([]byte, error) {
	chunk := make([]byte, length)
	read, err := e.r.Read(chunk)
	if uint32(read) != length {
		return nil, ErrIncomplete
	}
	if err != nil {
		return nil, err
	}
	return chunk, nil
}

// readLength reads the length-prefix tag from the reader and returns its
// uint32 value.
func (e *Decoder) readLength() (uint32, error) {
	binLength := make([]byte, lengthSpecSize)
	read, err := e.r.Read(binLength)
	if read == 0 && err == io.EOF {
		return 0, err
	}
	if read != lengthSpecSize {
		return 0, ErrIncomplete
	}
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(binLength), nil
}
