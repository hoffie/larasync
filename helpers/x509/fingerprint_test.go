package x509

import (
	"path/filepath"

	. "gopkg.in/check.v1"
)

type FingerprintTests struct{}

var _ = Suite(&FingerprintTests{})

func (t *FingerprintTests) TestPEM(c *C) {
	out := c.MkDir()
	keyFile := filepath.Join(out, "lara-server.key")
	certFile := filepath.Join(out, "lara-server.crt")
	err := GenerateServerCertFiles(certFile, keyFile)
	c.Assert(err, IsNil)
	fp, err := CertificateFingerprintFromPEMFile(certFile)
	c.Assert(err, IsNil)
	c.Assert(len(fp), Equals, 128)
}
