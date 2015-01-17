package x509

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"os"
	"time"
)

const (
	certFileName = "lara-server.crt"
	keyFileName  = "lara-server.key"
)

// GenerateServerCert generates a new server certificate, writing the resulting
// keys and certificates to the provided writers.
func GenerateServerCert(keyOut, certOut io.Writer) error {
	// inspired by crypto/tls/generate_cert.go
	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return err
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(128 * 365 * 24 * time.Hour)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"larasync server certificate"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template,
		&priv.PublicKey, priv)
	if err != nil {
		return err
	}

	pemBlock, err := pemBlockForKey(priv)
	if err != nil {
		return err
	}
	err = pem.Encode(keyOut, pemBlock)
	if err != nil {
		return err
	}

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return err
	}

	return nil
}

// pemBlockForKey returns a pem block containing the given private key.
func pemBlockForKey(k *ecdsa.PrivateKey) (*pem.Block, error) {
	b, err := x509.MarshalECPrivateKey(k)
	if err != nil {
		return nil, err
	}
	return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}, nil
}

// GenerateServerCertFiles creates a certificate and a key file in the provided
// output directory.
func GenerateServerCertFiles(certFile, keyFile string) error {
	certOut, err := os.OpenFile(certFile, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer certOut.Close()

	keyOut, err := os.OpenFile(keyFile, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	return GenerateServerCert(keyOut, certOut)
}
