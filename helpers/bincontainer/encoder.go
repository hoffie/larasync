package bincontainer

import (
	"encoding/binary"
	"io"
)

// Encoder can write arbitrary binary chunks into a Writer in a way that they can be
// separated again later by a Decoder.
// NOTE: a maximum length of 2**32 (~4GB) is supported.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns a new encoder reference.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// WriteChunk forwards the given bytes to the underlying writer, prefixed by the
// length.
func (e *Encoder) WriteChunk(chunk []byte) error {
	err := e.writeLength(len(chunk))
	if err != nil {
		return err
	}
	_, err = e.w.Write(chunk)
	return err
}

// writeLength writes the given length to the underlying writer.
func (e *Encoder) writeLength(length int) error {
	binLength := make([]byte, lengthSpecSize)
	binary.LittleEndian.PutUint32(binLength, uint32(length))

	_, err := e.w.Write(binLength)
	return err
}
