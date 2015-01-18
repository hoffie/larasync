package api

import (
	"fmt"

	"crypto/rand"
	"encoding/hex"

	. "gopkg.in/check.v1"

	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
	"github.com/hoffie/larasync/repository"
)

type AuthorizationTests struct {
	BaseServerTest
	encryptionKey  [repository.EncryptionKeySize]byte
	authPrivateKey [repository.PrivateKeySize]byte
	authPublicKey  [repository.PublicKeySize]byte
}

func getAuthorizationTest() AuthorizationTests {
	return AuthorizationTests{
		BaseServerTest: newBaseServerTest(),
	}
}

func (t *AuthorizationTests) SetUpTest(c *C) {
	t.BaseServerTest.SetUpTest(c)
	t.encryptionKey = [repository.EncryptionKeySize]byte{}
	t.authPrivateKey = [repository.PrivateKeySize]byte{}

	rand.Read(t.encryptionKey[:])
	byteArray := make([]byte, 200)
	rand.Read(byteArray)
	var err error
	t.authPrivateKey, err = passphraseToKey(byteArray)
	c.Assert(err, IsNil)
	t.authPublicKey = edhelpers.GetPublicKeyFromPrivate(t.authPrivateKey)

	getURL := t.getURL
	t.getURL = func() string {
		return fmt.Sprintf(
			"%s/authorizations/%s",
			getURL(),
			hex.EncodeToString(t.authPublicKey[:]),
		)
	}
	t.req = t.requestEmptyBody(c)
}

func (t *AuthorizationTests) testAuthorization(c *C) *repository.Authorization {
	return &repository.Authorization{}
}

func (t *AuthorizationTests) addAuthorization(c *C, auth *repository.Authorization) {
	repository := t.getRepository(c)

	err := repository.SetAuthorization(t.authPublicKey, t.encryptionKey, auth)
	c.Assert(err, IsNil)
}
