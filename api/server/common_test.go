package server

import (
	"testing"

	"github.com/hoffie/larasync/api/common"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

var adminSecret = []byte("foo")
var adminPubkey [PublicKeySize]byte

func init() {
	var err error
	adminPubkey, err = common.GetAdminSecretPubkey(adminSecret)
	if err != nil {
		panic(err)
	}
}
