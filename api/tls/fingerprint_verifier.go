package tls

import (
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"net"
	"time"
)

// ErrFingerprintRejected is returned when the fingerprint cannot be verified.
var ErrFingerprintRejected = errors.New("fingerprint rejected")

type handshakeTimeoutError struct{}

func (handshakeTimeoutError) Timeout() bool   { return true }
func (handshakeTimeoutError) Temporary() bool { return true }
func (handshakeTimeoutError) Error() string {
	return "api/tls: TLS handshake timeout"
}

const handshakeTimeout = 10 * time.Second

// VerificationFunc is the interface all callbacks have to fullfill so that they
// can act as a fingerprint verifier.
type VerificationFunc func(string) bool

// FingerprintVerifier provides a TLS connection handler which validates connections
// based on the server's fingerprint.
type FingerprintVerifier struct {
	AcceptFingerprint string
	VerificationFunc  VerificationFunc
}

// DialTLS is the function which hooks into net/http.Transport and should be
// passed as a function reference.
func (v *FingerprintVerifier) DialTLS(network, addr string) (net.Conn, error) {
	// setting InsecureSkipVerify here so that net/tls does not perform any
	// validations; we validate the certificate fingerprint later.
	cfg := &tls.Config{InsecureSkipVerify: true}
	plainConn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	tlsConn := tls.Client(plainConn, cfg)
	errc := make(chan error, 2)

	// for canceling TLS handshake
	timer := time.AfterFunc(handshakeTimeout, func() {
		errc <- handshakeTimeoutError{}
	})
	go func() {
		err := tlsConn.Handshake()
		if timer != nil {
			timer.Stop()
		}
		errc <- err
	}()
	if err := <-errc; err != nil {
		plainConn.Close()
		return nil, err
	}
	// note: this is the place where hostname verification usually occurs;
	// we do not do this as we do not use a CA infrastructure;
	// instead, we do fingerprint verification here
	cs := tlsConn.ConnectionState()
	if len(cs.PeerCertificates) < 1 {
		plainConn.Close()
		return nil, err
	}
	peerCert := cs.PeerCertificates[0]
	if !v.acceptPeerCert(peerCert) {
		plainConn.Close()
		return nil, ErrFingerprintRejected
	}
	return tlsConn, err
}

// acceptPeerCert decides whether the given cert is accepted; it first checks
// if the fingerprint is white-listed already; if it isn't, the verification
// callback is invoked if non-nil.
// in all other cases the certificate is rejected
func (v *FingerprintVerifier) acceptPeerCert(cert *x509.Certificate) bool {
	fp := CertificateFingerprint(cert)
	if v.AcceptFingerprint != "" {
		return v.AcceptFingerprint == fp
	}
	if v.VerificationFunc == nil {
		return false
	}
	res := v.VerificationFunc(fp)
	if res {
		v.AcceptFingerprint = fp
	}
	return res
}

// CertificateFingerprint returns the SHA-512 fingerprint of the given certificate
// as a hex-encoded string.
func CertificateFingerprint(cert *x509.Certificate) string {
	h := sha512.New()
	h.Write(cert.RawSubjectPublicKeyInfo)
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}
