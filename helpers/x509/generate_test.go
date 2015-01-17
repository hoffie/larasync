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
	err := GenerateServerCertFiles(out)
	c.Assert(err, IsNil)
	_, err = os.Stat(filepath.Join(out, "lara-server.key"))
	c.Assert(err, IsNil)
	_, err = os.Stat(filepath.Join(out, "lara-server.crt"))
	c.Assert(err, IsNil)
}
