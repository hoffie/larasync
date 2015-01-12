package api

import (
	"bytes"
	"io/ioutil"
	"net"
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

func (t *SignTests) TestAdminSigningIgnoreUserAgent(c *C) {
	t.req.Header.Set("User-Agent", "foo")
	c.Assert(t.adminSigned(), Equals, true)
}

// TestSigningAvoidHeaderMixup verifies that different parts
// of the request may not be confused with other parts by the
// signature algorithm.
func (t *SignTests) TestSigningAvoidHeaderMixup(c *C) {
	t.req.Header.Set("Header", "value")
	SignWithPassphrase(t.req, adminSecret)
	c.Assert(t.adminSigned(), Equals, true)
	t.req.Header.Del("Header")
	t.req.Header.Set("Headerval", "ue")
	c.Assert(t.adminSigned(), Equals, false)
}

func (t *SignTests) TestAdminSigningIgnoreHost(c *C) {
	// we ignore the Host header as it breaks signing due to
	// differences in client-side and server-side requests;
	// the actual host name is still signed as part of the URL
	t.req.Header.Set("Host", "foo")
	c.Assert(t.adminSigned(), Equals, true)
}

func (t *SignTests) TestAdminSigningIgnoreAcceptEncoding(c *C) {
	t.req.Header.Set("Accept-Encoding", "foo")
	c.Assert(t.adminSigned(), Equals, true)
}

func (t *SignTests) TestAdminSigningNormalizeURL(c *C) {
	t.req.URL.Host = ""
	t.req.Host = "example.org"
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

// TestRealRoundTrip uses the full Go http client/server stack to execute
// a real HTTP roundtrip.
// This ensures that this process does not mangle the request in a way
// which would break signatures.
func (t *SignTests) TestRealRoundTrip(c *C) {
	// passing port :0 to Listen lets it choose a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	c.Assert(err, IsNil)
	defer listener.Close()
	hostAndPort := listener.Addr().String()

	adminSecret := []byte("test")
	pubKey, err := GetAdminSecretPubkey(adminSecret)
	c.Assert(err, IsNil)

	server := &http.Server{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if !ValidateRequest(req, pubKey, 5*time.Second) {
				rw.WriteHeader(http.StatusUnauthorized)
				return
			}
			rw.Header().Set("X-Lara-Validated", "1")
		}),
	}
	go server.Serve(listener)

	body := bytes.NewReader([]byte("test"))
	req, err := http.NewRequest("GET", "http://"+hostAndPort+"/foo.txt?x=1", body)
	c.Assert(err, IsNil)
	SignWithPassphrase(req, adminSecret)
	client := &http.Client{}
	resp, err := client.Do(req)
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, 200)
	c.Assert(resp.Header.Get("X-Lara-Validated"), Equals, "1")
}
