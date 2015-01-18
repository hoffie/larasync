package main

import . "gopkg.in/check.v1"

type ServerFingerprintTests struct {
	BaseTests
}

var _ = Suite(&ServerFingerprintTests{})

func (t *ServerFingerprintTests) TestFail(c *C) {
	c.Assert(t.d.run([]string{"server-fingerprint"}), Equals, 1)
}

func (t *ServerFingerprintTests) Test(c *C) {
	err := t.d.needServerCert()
	c.Assert(err, IsNil)
	c.Assert(t.d.run([]string{"server-fingerprint"}), Equals, 0)
	out := t.out.String()
	// as we output a colored hash, the actual length is longer, which we don't
	// validate here
	c.Assert(len(out) >= 129 /*hex(SHA512) + '\n'*/, Equals, true)
}
