package main

import (
	"github.com/hoffie/larasync/repository"

	. "gopkg.in/check.v1"
)

type ResetFingerprintTests struct {
	BaseTests
	sc *repository.StateConfig
}

var _ = Suite(&ResetFingerprintTests{})

func (t *ResetFingerprintTests) SetUpTest(c *C) {
	t.BaseTests.SetUpTest(c)
	c.Assert(t.d.run([]string{"init"}), Equals, 0)

	r, err := repository.NewClient(t.dir)
	c.Assert(err, IsNil)
	_ = r
	t.sc, err = r.StateConfig()
	c.Assert(err, IsNil)
	t.sc.DefaultServer.Fingerprint = "123"
	err = t.sc.Save()
	c.Assert(err, IsNil)
}

func (t *ResetFingerprintTests) TestNo(c *C) {
	t.in.WriteString("n\n")
	c.Assert(t.d.run([]string{"reset-fingerprint"}), Equals, 0)

	err := t.sc.Load()
	c.Assert(err, IsNil)
	c.Assert(t.sc.DefaultServer.Fingerprint, Equals, "123")
}

func (t *ResetFingerprintTests) TestYes(c *C) {
	t.in.WriteString("y\n")
	c.Assert(t.d.run([]string{"reset-fingerprint"}), Equals, 0)

	err := t.sc.Load()
	c.Assert(err, IsNil)
	c.Assert(t.sc.DefaultServer.Fingerprint, Equals, "")
}
