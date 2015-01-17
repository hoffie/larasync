package x509

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	. "gopkg.in/check.v1"
)

type GenerateTests struct{}

var _ = Suite(&GenerateTests{})

func (t *GenerateTests) TestGenerateServerCert(c *C) {
	keyOut := &bytes.Buffer{}
	certOut := &bytes.Buffer{}
	err := GenerateServerCert(keyOut, certOut)
	c.Assert(err, IsNil)
	c.Assert(strings.Index(keyOut.String(), "EC PRIVATE KEY"), Not(Equals), -1)
	c.Assert(strings.Index(certOut.String(), "CERTIFICATE"), Not(Equals), -1)
}

func (t *GenerateTests) TestGenerateFiles(c *C) {
	out := c.MkDir()
	keyFile := filepath.Join(out, "lara-server.key")
	certFile := filepath.Join(out, "lara-server.crt")
	err := GenerateServerCertFiles(certFile, keyFile)
	c.Assert(err, IsNil)
	_, err = os.Stat(keyFile)
	c.Assert(err, IsNil)
	_, err = os.Stat(certFile)
	c.Assert(err, IsNil)
}
