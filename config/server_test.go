package config

import (
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type ConfigSanitizeTests struct{}

var _ = Suite(&ConfigSanitizeTests{})

func (t *ConfigSanitizeTests) TestListen(c *C) {
	sc := &ServerConfig{}
	sc.Sanitize()
	c.Assert(sc.Server.Listen, Equals, "127.0.0.1:14124")
}

func (t *ConfigSanitizeTests) TestAdminPubkeyMissing(c *C) {
	sc := &ServerConfig{}
	err := sc.Sanitize()
	c.Assert(err, Equals, ErrAdminPubkeyMissing)
}

func (t *ConfigSanitizeTests) TestAdminPubkeyBad(c *C) {
	sc := &ServerConfig{}
	sc.Signatures.AdminPubkey = "foo"
	err := sc.Sanitize()
	c.Assert(err, Equals, ErrInvalidAdminPubkey)
}

func (t *ConfigSanitizeTests) TestAdminPubkeyTooShort(c *C) {
	sc := &ServerConfig{}
	sc.Signatures.AdminPubkey = "1234"
	err := sc.Sanitize()
	c.Assert(err, Equals, ErrTruncatedAdminPubkey)
}

func (t *ConfigSanitizeTests) TestMissingBasePath(c *C) {
	sc := &ServerConfig{}
	sc.Signatures.AdminPubkey = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	err := sc.Sanitize()
	c.Assert(err, Equals, ErrMissingBasePath)
}

func (t *ConfigSanitizeTests) TestBadBasePath(c *C) {
	sc := &ServerConfig{}
	sc.Signatures.AdminPubkey = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	sc.Repository.BasePath = "/dev/null"
	err := sc.Sanitize()
	c.Assert(err, Equals, ErrBadBasePath)
}

func (t *ConfigSanitizeTests) TestOk(c *C) {
	dir := c.MkDir()
	sc := &ServerConfig{}
	sc.Signatures.AdminPubkey = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	sc.Repository.BasePath = dir
	err := sc.Sanitize()
	c.Assert(err, IsNil)
}

func (t *ConfigSanitizeTests) TestSignatureMaxAge(c *C) {
	dir := c.MkDir()
	sc := &ServerConfig{}
	sc.Signatures.AdminPubkey = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	sc.Repository.BasePath = dir
	err := sc.Sanitize()
	c.Assert(err, IsNil)
	c.Assert(sc.Signatures.MaxAge, Equals, 5*time.Second)
}
