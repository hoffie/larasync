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
	SignAsAdmin(req, adminSecret)
	t.req = req
}

func (t *SignTests) adminSigned() bool {
	return ValidateAdminSigned(t.req, adminSecret, time.Minute)
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

func (t *SignTests) TestAdminSigningNonAdminAuthorizationHeader(c *C) {
	t.req.Header.Set("Authorization", "lara foo")
	c.Assert(t.adminSigned(), Equals, false)
}

func (t *SignTests) TestAdminSigningNonGivenAuthorizationHeader(c *C) {
	t.req.Header.Set("Authorization", "lara admin ")
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
	SignAsAdmin(t.req, adminSecret)
	c.Assert(ValidateAdminSigned(t.req, adminSecret, 9*time.Second), Equals, false)
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
