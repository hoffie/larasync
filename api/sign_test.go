package api

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	. "gopkg.in/check.v1"
)

type SignTests struct {
	req *http.Request
}

var _ = Suite(&SignTests{})

func (t *SignTests) SetUpTest(c *C) {
	req, err := http.NewRequest("GET", "http://example.org/repositories", nil)
	c.Assert(err, IsNil)
	SignWithPassphrase(req, adminSecret)
	t.req = req
}

func (t *SignTests) adminSigned() bool {
	return ValidateRequest(t.req, adminPubkey, time.Minute)
}

func (t *SignTests) TestAuthorizationHeader(c *C) {
	c.Assert(t.req.Header.Get("Authorization"), Not(Equals), "")
}

func (t *SignTests) TestAdminSigningCorrectSignature(c *C) {
	c.Assert(t.adminSigned(), Equals, true)
}

func (t *SignTests) TestAdminSigningEmptyAuthorizationHeader(c *C) {
	t.req.Header.Set("Authorization", "")
	c.Assert(t.adminSigned(), Equals, false)
}

func (t *SignTests) TestAdminSigningNonLaraAuthorizationHeader(c *C) {
	t.req.Header.Set("Authorization", "basic foo")
	c.Assert(t.adminSigned(), Equals, false)
}

func (t *SignTests) TestAdminSigningShortSig(c *C) {
	t.req.Header.Set("Authorization", "lara 111")
	c.Assert(t.adminSigned(), Equals, false)
}

func (t *SignTests) TestAdminSigningMissingAuthorizationSig(c *C) {
	t.req.Header.Set("Authorization", "lara ")
	c.Assert(t.adminSigned(), Equals, false)
}

func (t *SignTests) TestAdminSigningBadHexHash(c *C) {
	t.req.Header.Set("Authorization", "lara 123")
	c.Assert(t.adminSigned(), Equals, false)
}

func (t *SignTests) TestAdminSigningHashTooShort(c *C) {
	t.req.Header.Set("Authorization", "lara 1234")
	c.Assert(t.adminSigned(), Equals, false)
}

func (t *SignTests) TestAdminSigningChangedMethodAuthorizationHeader(c *C) {
	t.req.Method = "POST"
	c.Assert(t.adminSigned(), Equals, false)
}

func (t *SignTests) TestAdminSigningAddedHeader(c *C) {
	t.req.Header.Set("Test", "1")
	c.Assert(t.adminSigned(), Equals, false)
}

func (t *SignTests) TestAdminSigningChangedUrl(c *C) {
	newURL, err := url.Parse("http://example.org/repositories?x=3")
	c.Assert(err, IsNil)
	t.req.URL = newURL
	c.Assert(t.adminSigned(), Equals, false)
}

func (t *SignTests) TestAdminSigningOutdatedSignature(c *C) {
	tenSecsAgo := time.Now().Add(-10 * time.Second)
	// this will most likely send non-GMT time which is against the HTTP RFC;
	// as we should handle this as well, it's ok for testing:
	t.req.Header.Set("Date", tenSecsAgo.Format(time.RFC1123))
	SignWithPassphrase(t.req, adminSecret)
	c.Assert(ValidateRequest(t.req, adminPubkey, 9*time.Second), Equals, false)
}

func (t *SignTests) changeBody() {
	t.req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("changed body")))
}

func (t *SignTests) TestAdminSigningBodyChange(c *C) {
	t.changeBody()
	c.Assert(t.adminSigned(), Equals, false)
}

func (t *SignTests) TestAdminSigningBodyTextRead(c *C) {
	t.changeBody()
	c.Assert(t.adminSigned(), Equals, false)
	buf := make([]byte, 100)
	read, _ := t.req.Body.Read(buf)
	c.Assert(string(buf[:read]), Equals, "changed body")
}

func (t *SignTests) TestYoungerThanBadHeader(c *C) {
	t.req.Header.Set("Date", "123")
	c.Assert(youngerThan(t.req, time.Minute), Equals, false)
}
