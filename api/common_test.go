package api

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

var adminSecret = []byte("foo")
var adminPubkey [PublicKeySize]byte

func init() {
	var err error
	adminPubkey, err = GetAdminSecretPubkey(adminSecret)
	if err != nil {
		panic(err)
	}
}
