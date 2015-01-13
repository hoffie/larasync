package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"

	. "gopkg.in/check.v1"

	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
)

type AuthorizationURLTests struct {
	encKey  [EncryptionKeySize]byte
	signKey [PrivateKeySize]byte
	pubKey  [PublicKeySize]byte
	baseURL string
}

var _ = Suite(&AuthorizationURLTests{})

func (t *AuthorizationURLTests) SetUpTest(c *C) {
	pubKey, signKey, _ := edhelpers.GenerateKey()
	t.pubKey = *pubKey
	t.signKey = *signKey
	rand.Read(t.encKey[:])
	t.baseURL = "http://example.org/repo"
}

func (t *AuthorizationURLTests) getAuthURL() *AuthorizationURL {
	auth, _ := newAuthURL(t.baseURL, &t.signKey, &t.encKey)
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
		"%s#AuthEncKey=%s&AuthSignKey=%s",
		t.getTestRequestURLString(),
		hex.EncodeToString(t.encKey[:]),
		hex.EncodeToString(t.signKey[:]),
	)
}

func (t *AuthorizationURLTests) TestGenerate(c *C) {
	authURL := t.getAuthURL()
	c.Assert(authURL.SignKey, DeepEquals, t.signKey)
	c.Assert(authURL.EncKey, DeepEquals, t.encKey)

	c.Assert(authURL.String(), Equals, t.getTestURLString())
}

func (t *AuthorizationURLTests) TestParse(c *C) {
	u, err := url.Parse(t.getTestURLString())
	c.Assert(err, IsNil)
	authURL, err := parseAuthURL(u)
	c.Assert(err, IsNil)
	c.Assert(authURL.SignKey, DeepEquals, t.signKey)
	c.Assert(authURL.EncKey, DeepEquals, t.encKey)
	c.Assert(authURL.URL.String(), Equals, t.getTestRequestURLString())
}

func (t *AuthorizationURLTests) TestGenerateUrl(c *C) {
	_, err := newAuthURL("%(asdf", &t.signKey, &t.encKey)
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
			"%s#AuthSignKey=%s",
			t.getTestRequestURLString(),
			hex.EncodeToString(t.encKey[:]),
		),
	)
	c.Assert(err, IsNil)
	_, err = parseAuthURL(u)
	c.Assert(err, NotNil)
}

func (t *AuthorizationURLTests) TestToShortSignKey(c *C) {
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

func (t *AuthorizationURLTests) TestToShortSignKeyEncodingError(c *C) {
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

func (t *AuthorizationURLTests) TestToShortEncryptionKey(c *C) {
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

func (t *AuthorizationURLTests) TestToShortEncryptionKeyEncodingError(c *C) {
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
