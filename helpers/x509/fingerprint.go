package x509

import (
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"io/ioutil"
)

// CertificateFingerprint returns the SHA-512 fingerprint of the given certificate
// as a hex-encoded string.
//
// Note that this algorithm is *not* compatible with the one used in browsers.
// This is for two reasons:
// Firstly, currentl (2015) browsers do not show sha512 fingerprints anway
// (sha256 and sha1 only).
// Secondly, we do not care about the certificate at all (it contains standardized
// field contents only). Instead we want to verify the public key. This is what
// we do here.
//
// The algorithm is taken from
// http://tools.ietf.org/html/draft-ietf-websec-key-pinning-01#ref-why-pin-key
// but uses sha512 instead of sha1.
func CertificateFingerprint(cert *x509.Certificate) string {
	h := sha512.New()
	h.Write(cert.RawSubjectPublicKeyInfo)
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

// CertificateFingerprintFromBytes parses the given certificate bytes and returns its
// fingerprint.
func CertificateFingerprintFromBytes(cert []byte) (string, error) {
	parsed, err := x509.ParseCertificate(cert)
	if err != nil {
		return "", err
	}
	return CertificateFingerprint(parsed), nil
}

// CertificateFingerprintFromPEMFile loads the given PEM file and returns its
// fingerprint.
func CertificateFingerprintFromPEMFile(path string) (string, error) {
	pemBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return "", errors.New("unable to parse PEM block")
	}
	return CertificateFingerprintFromBytes(block.Bytes)
}
