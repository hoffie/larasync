package repository

import (
	"crypto/sha512"
	"errors"
	"hash"
	"io"
	"os"

	"github.com/agl/ed25519"
)

const (
	// PrivateKeySize denotes how many bytes a private key needs (binary encoded)
	PrivateKeySize = ed25519.PrivateKeySize
	// PublicKeySize denotes how many bytes a pubkey needs (binary encoded)
	PublicKeySize = ed25519.PublicKeySize
	// SignatureSize denotes how many bytes a sig needs (binary encoded)
	SignatureSize = ed25519.SignatureSize
)

// SigningWriter is Writer which secures written
// data with a signature.
type SigningWriter struct {
	hash    hash.Hash
	writer  io.Writer
	privKey [PrivateKeySize]byte
}

// NewSigningWriter returns a new SigningWriter instance.
func NewSigningWriter(key [PrivateKeySize]byte, writer io.Writer) *SigningWriter {
	s := &SigningWriter{
		hash:    sha512.New(),
		writer:  writer,
		privKey: key,
	}
	return s
}

// Write implements the Writer interface.
func (w *SigningWriter) Write(data []byte) (int, error) {
	written, err := w.writer.Write(data)
	if err != nil {
		return written, err
	}
	_, err = w.hash.Write(data[:written])
	return written, err
}

// Finalize writes the signature. No more Write() calls are allowed to
// happen afterwards (not enforced atm).
func (w *SigningWriter) Finalize() error {
	sum := w.hash.Sum(nil)
	sig := ed25519.Sign(&w.privKey, sum)
	if sig == nil {
		return errors.New("signing failed")
	}
	_, err := w.writer.Write(sig[:])
	return err
}

// VerifyingReader is a Reader which is able to Verify all data read
// against the given public key. .Verify() has to be called, no
// implicit verification is built in!
type VerifyingReader struct {
	hash          hash.Hash
	reader        io.Reader
	limitedReader io.Reader
	pubKey        [PublicKeySize]byte
	sig           [SignatureSize]byte
	dataLength    int64
}

// NewVerifyingReader returns a new VerifyingReader.
func NewVerifyingReader(pubKey [PublicKeySize]byte, reader io.ReadSeeker) (*VerifyingReader, error) {
	r := &VerifyingReader{
		hash:   sha512.New(),
		reader: reader,
		pubKey: pubKey,
	}
	var err error
	r.dataLength, err = reader.Seek(-SignatureSize, os.SEEK_END)
	if err != nil {
		return nil, err
	}
	r.limitedReader = &io.LimitedReader{
		R: r.reader,
		N: r.dataLength,
	}
	_, err = r.reader.Read(r.sig[:])
	if err != nil {
		return nil, err
	}
	_, err = reader.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Read implements the Reader interface.
func (r *VerifyingReader) Read(buf []byte) (int, error) {
	read, err := r.limitedReader.Read(buf)
	_, err2 := r.hash.Write(buf[:read])
	if err2 != nil {
		return read, err2
	}
	if err != nil {
		return read, err
	}
	return read, nil
}

// VerifyAfterRead returns whether the data read is ok (as in: signature
// matches and verifies with the given public key).
//
// IMPORTANT: This must only be called after *all data has been read*;
// otherwise verification will always fail.
func (r *VerifyingReader) VerifyAfterRead() bool {
	sum := r.hash.Sum(nil)
	return ed25519.Verify(&r.pubKey, sum, &r.sig)
}
