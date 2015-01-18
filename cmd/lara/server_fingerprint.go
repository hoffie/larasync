package main

import (
	"fmt"

	"github.com/hoffie/larasync/helpers/x509"
)

// serverFingerprintAction outputs the server's certificate public key fingerprint
func (d *Dispatcher) serverFingerprintAction() int {
	haveKeys, err := d.haveServerCert()
	if err != nil {
		fmt.Fprintf(d.stderr, "unable to check for server certificates (%s)\n", err)
		return 1
	}
	if !haveKeys {
		fmt.Fprintf(d.stderr, "no server certificate found; start the server at least once\n")
		return 1
	}
	certFile, _ := d.serverCertFilePaths()
	fp, err := x509.CertificateFingerprintFromPEMFile(certFile)
	fmt.Fprintf(d.stdout, "%s\n", fp)
	return 0
}
