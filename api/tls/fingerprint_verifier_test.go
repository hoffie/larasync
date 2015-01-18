package tls

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"math/big"
	"strings"

	. "gopkg.in/check.v1"
)

type FVTests struct{}

var _ = Suite(&FVTests{})

func (t *FVTests) TestAcceptPeerCertNoFpNoFunc(c *C) {
	crt := t.makeTestCert(c)
	fpv := FingerprintVerifier{}
	c.Assert(fpv.acceptPeerCert(crt), Equals, false)
}

func (t *FVTests) TestAcceptPeerCertAcceptFuncFail(c *C) {
	crt := t.makeTestCert(c)
	fpv := FingerprintVerifier{
		VerificationFunc: func(fp string) bool {
			return false
		},
	}
	c.Assert(fpv.acceptPeerCert(crt), Equals, false)
	c.Assert(fpv.AcceptFingerprint, Equals, "")
}

func (t *FVTests) TestAcceptPeerCertAcceptFuncAccept(c *C) {
	crt := t.makeTestCert(c)
	fpv := FingerprintVerifier{
		VerificationFunc: func(fp string) bool {
			return true
		},
	}
	c.Assert(fpv.acceptPeerCert(crt), Equals, true)
	c.Assert(fpv.AcceptFingerprint, Equals, testCertFp)
	fpv.VerificationFunc = func(fp string) bool {
		c.Fatal("unexpected call to verification func")
		return false
	}
	c.Assert(fpv.acceptPeerCert(crt), Equals, true)
}

func (t *FVTests) TestAcceptPeerCertPreset(c *C) {
	crt := t.makeTestCert(c)
	fpv := FingerprintVerifier{
		AcceptFingerprint: testCertFp,
	}
	c.Assert(fpv.acceptPeerCert(crt), Equals, true)
}

func (t *FVTests) TestAcceptPeerCertPresetFail(c *C) {
	crt := t.makeTestCert(c)
	fpv := FingerprintVerifier{
		AcceptFingerprint: strings.Replace(testCertFp, "c", "d", 1),
	}
	c.Assert(fpv.acceptPeerCert(crt), Equals, false)
}

const testCertFp = "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"

func (t *FVTests) makeTestCert(c *C) *x509.Certificate {
	entropy := bytes.NewBufferString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	priv, err := ecdsa.GenerateKey(elliptic.P521(), entropy)
	c.Assert(err, IsNil)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
	}

	_, err = x509.CreateCertificate(entropy, &template, &template,
		&priv.PublicKey, priv)
	c.Assert(err, IsNil)
	return &template
}
