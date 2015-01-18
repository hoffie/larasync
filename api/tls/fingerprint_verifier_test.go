package tls

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"math/big"
	"net"
	"strings"

	. "gopkg.in/check.v1"
)

type FVTests struct{}

var _ = Suite(&FVTests{})

func (t *FVTests) TestAcceptPeerCertNoFpNoFunc(c *C) {
	crt, _, _ := t.makeTestCert(c)
	fpv := FingerprintVerifier{}
	c.Assert(fpv.acceptPeerCert(crt), Equals, false)
}

func (t *FVTests) TestAcceptPeerCertAcceptFuncFail(c *C) {
	crt, _, _ := t.makeTestCert(c)
	fpv := FingerprintVerifier{
		VerificationFunc: func(fp string) bool {
			return false
		},
	}
	c.Assert(fpv.acceptPeerCert(crt), Equals, false)
	c.Assert(fpv.AcceptFingerprint, Equals, "")
}

func (t *FVTests) TestAcceptPeerCertAcceptFuncAccept(c *C) {
	crt, _, _ := t.makeTestCert(c)
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
	crt, _, _ := t.makeTestCert(c)
	fpv := FingerprintVerifier{
		AcceptFingerprint: testCertFp,
	}
	c.Assert(fpv.acceptPeerCert(crt), Equals, true)
}

func (t *FVTests) TestAcceptPeerCertPresetFail(c *C) {
	crt, _, _ := t.makeTestCert(c)
	fpv := FingerprintVerifier{
		AcceptFingerprint: strings.Replace(testCertFp, "c", "d", 1),
	}
	c.Assert(fpv.acceptPeerCert(crt), Equals, false)
}

func (t *FVTests) TestAcceptPeerCertPresetFailNoCall(c *C) {
	crt, _, _ := t.makeTestCert(c)
	fpv := FingerprintVerifier{
		AcceptFingerprint: strings.Replace(testCertFp, "c", "d", 1),
		VerificationFunc: func(fp string) bool {
			c.Fatal("verification called although unexpected")
			return false
		},
	}
	c.Assert(fpv.acceptPeerCert(crt), Equals, false)
}

func (t *FVTests) setupTLSServer(c *C) (string, net.Listener) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	c.Assert(err, IsNil)
	_, derBytes, priv := t.makeTestCert(c)
	myCert := tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  priv,
	}
	tlsL := tls.NewListener(l, &tls.Config{
		NextProtos:   []string{"http/1.1"},
		Certificates: []tls.Certificate{myCert},
	})
	go func() {
		conn, err := tlsL.Accept()
		c.Assert(err, IsNil)
		defer conn.Close()
		_, err = conn.Write([]byte("test"))
		c.Assert(err, IsNil)
	}()
	addr := tlsL.Addr().String()
	return addr, l
}

func (t *FVTests) TestDialTLSReject(c *C) {
	addr, l := t.setupTLSServer(c)
	defer l.Close()
	fpv := &FingerprintVerifier{}
	_, err := fpv.DialTLS("tcp", addr)
	c.Assert(err, Equals, ErrFingerprintRejected)
}

func (t *FVTests) TestDialTLSAccept(c *C) {
	addr, l := t.setupTLSServer(c)
	defer l.Close()
	fpv := &FingerprintVerifier{AcceptFingerprint: testCertFp}
	conn, err := fpv.DialTLS("tcp", addr)
	c.Assert(err, IsNil)
	r, err := ioutil.ReadAll(conn)
	c.Assert(err, IsNil)
	c.Assert(r, DeepEquals, []byte("test"))
}

const testCertFp = "9f79df7d821ea16d89c09e026074e81f89540aa7fbfda1b0b3f5ba7dcab88d71b944a49bb0c8a5c8abf42308d8ae060bf7437831e7a5f21b3c7718f04578680b"

func (t *FVTests) makeTestCert(c *C) (*x509.Certificate, []byte, *ecdsa.PrivateKey) {
	entropy := bytes.NewBufferString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	priv, err := ecdsa.GenerateKey(elliptic.P521(), entropy)
	c.Assert(err, IsNil)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
	}

	derBytes, err := x509.CreateCertificate(entropy, &template, &template,
		&priv.PublicKey, priv)
	c.Assert(err, IsNil)
	cert, err := x509.ParseCertificate(derBytes)
	c.Assert(err, IsNil)
	return cert, derBytes, priv
}
