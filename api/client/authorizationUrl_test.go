package client

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"

	. "gopkg.in/check.v1"

	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
	"github.com/hoffie/larasync/repository"
)

type AuthorizationURLTests struct {
	encKey      [repository.EncryptionKeySize]byte
	signKey     [PrivateKeySize]byte
	pubKey      [PublicKeySize]byte
	fingerprint string
	baseURL     string
}

var _ = Suite(&AuthorizationURLTests{})

func (t *AuthorizationURLTests) SetUpTest(c *C) {
	pubKey, signKey, _ := edhelpers.GenerateKey()
	t.pubKey = *pubKey
	t.signKey = *signKey
	t.fingerprint = "test"
	rand.Read(t.encKey[:])
	t.baseURL = "https://example.org/repo"
}

func (t *AuthorizationURLTests) getAuthURL() *AuthorizationURL {
	auth, _ := NewAuthURL(t.baseURL, &t.signKey, &t.encKey, t.fingerprint)
	return auth
}

func (t *AuthorizationURLTests) getTestRequestURLString() string {
	return fmt.Sprintf(
		"%s/authorizations/%s",
		t.baseURL,
		hex.EncodeToString(t.pubKey[:]),
	)
}

func (t *AuthorizationURLTests) getTestURLString() string {
	return fmt.Sprintf(
		"%s#AuthEncKey=%s&AuthSignKey=%s&Fingerprint=%s",
		t.getTestRequestURLString(),
		hex.EncodeToString(t.encKey[:]),
		hex.EncodeToString(t.signKey[:]),
		t.fingerprint,
	)
}

func (t *AuthorizationURLTests) TestGenerate(c *C) {
	authURL := t.getAuthURL()
	c.Assert(authURL.SignKey, DeepEquals, t.signKey)
	c.Assert(authURL.EncKey, DeepEquals, t.encKey)
	c.Assert(authURL.Fingerprint, DeepEquals, t.fingerprint)

	c.Assert(authURL.String(), Equals, t.getTestURLString())
}

func (t *AuthorizationURLTests) TestParse(c *C) {
	u, err := url.Parse(t.getTestURLString())
	c.Assert(err, IsNil)
	authURL, err := parseAuthURL(u)
	c.Assert(err, IsNil)
	c.Assert(authURL.SignKey, DeepEquals, t.signKey)
	c.Assert(authURL.EncKey, DeepEquals, t.encKey)
	c.Assert(authURL.Fingerprint, DeepEquals, t.fingerprint)
	c.Assert(authURL.URL.String(), Equals, t.getTestRequestURLString())
}

func (t *AuthorizationURLTests) TestGenerateUrl(c *C) {
	_, err := NewAuthURL("%(asdf", &t.signKey, &t.encKey, t.fingerprint)
	c.Assert(err, NotNil)
}

func (t *AuthorizationURLTests) TestNoEncKey(c *C) {
	u, err := url.Parse(
		fmt.Sprintf(
			"%s#AuthSignKey=%s",
			t.getTestRequestURLString(),
			hex.EncodeToString(t.signKey[:]),
		),
	)
	c.Assert(err, IsNil)
	_, err = parseAuthURL(u)
	c.Assert(err, NotNil)
}

func (t *AuthorizationURLTests) TestNoSignKey(c *C) {
	u, err := url.Parse(
		fmt.Sprintf(
			"%s#AuthEncKey=%s",
			t.getTestRequestURLString(),
			hex.EncodeToString(t.encKey[:]),
		),
	)
	c.Assert(err, IsNil)
	_, err = parseAuthURL(u)
	c.Assert(err, NotNil)
}

func (t *AuthorizationURLTests) TestNoFingerprint(c *C) {
	u, err := url.Parse(
		fmt.Sprintf(
			"%s#AuthSignKey=%s&AuthEncKey=%s",
			t.getTestRequestURLString(),
			hex.EncodeToString(t.signKey[:]),
			hex.EncodeToString(t.encKey[:]),
		),
	)
	c.Assert(err, IsNil)
	_, err = parseAuthURL(u)
	c.Assert(err, NotNil)
}

func (t *AuthorizationURLTests) TestTooShortSignKey(c *C) {
	signKeyString := hex.EncodeToString(t.signKey[:])
	u, err := url.Parse(
		fmt.Sprintf(
			"%s#AuthEncKey=%s&AuthSignKey=%s",
			t.getTestRequestURLString(),
			hex.EncodeToString(t.encKey[:]),
			signKeyString[:len(signKeyString)-2],
		),
	)
	c.Assert(err, IsNil)
	_, err = parseAuthURL(u)
	c.Assert(err, NotNil)
}

func (t *AuthorizationURLTests) TestTooShortSignKeyEncodingError(c *C) {
	signKeyString := hex.EncodeToString(t.signKey[:])
	u, err := url.Parse(
		fmt.Sprintf(
			"%s#AuthEncKey=%s&AuthSignKey=%s",
			t.getTestRequestURLString(),
			hex.EncodeToString(t.encKey[:]),
			signKeyString[:len(signKeyString)-1],
		),
	)
	c.Assert(err, IsNil)
	_, err = parseAuthURL(u)
	c.Assert(err, NotNil)
}

func (t *AuthorizationURLTests) TestTooShortEncryptionKey(c *C) {
	signKeyString := hex.EncodeToString(t.signKey[:])
	encKeyString := hex.EncodeToString(t.encKey[:])
	u, err := url.Parse(
		fmt.Sprintf(
			"%s#AuthEncKey=%s&AuthSignKey=%s",
			t.getTestRequestURLString(),
			encKeyString[:len(encKeyString)-2],
			signKeyString,
		),
	)
	c.Assert(err, IsNil)
	_, err = parseAuthURL(u)
	c.Assert(err, NotNil)
}

func (t *AuthorizationURLTests) TestTooShortEncryptionKeyEncodingError(c *C) {
	signKeyString := hex.EncodeToString(t.signKey[:])
	encKeyString := hex.EncodeToString(t.encKey[:])
	u, err := url.Parse(
		fmt.Sprintf(
			"%s#AuthEncKey=%s&AuthSignKey=%s",
			t.getTestRequestURLString(),
			encKeyString[:len(encKeyString)-1],
			signKeyString,
		),
	)
	c.Assert(err, IsNil)
	_, err = parseAuthURL(u)
	c.Assert(err, NotNil)
}
