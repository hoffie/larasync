package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	. "gopkg.in/check.v1"
)

func newBaseServerTest() BaseServerTest {
	return BaseServerTest{
		BaseTests: BaseTests{},
	}
}

type BaseServerTest struct {
	BaseTests
	server     *Server
	req        *http.Request
	httpMethod string
	getURL     func() string
	urlParams  url.Values
}

func (t *BaseServerTest) SetUpTest(c *C) {
	t.BaseTests.SetUpTest(c)
	t.createServer(c)
	t.httpMethod = "GET"
	t.getURL = func() string {
		return fmt.Sprintf(
			"http://example.org/repositories/%s",
			t.repositoryName,
		)
	}
	t.req = t.requestEmptyBody(c)
	t.urlParams = url.Values{}
}

func (t *BaseServerTest) createServer(c *C) {
	var err error
	t.server, err = New(adminPubkey, time.Minute, t.rm, t.certFile, t.keyFile)
	c.Assert(err, IsNil)
}

func (t *BaseServerTest) getResponse(req *http.Request) *httptest.ResponseRecorder {
	rw := httptest.NewRecorder()
	t.server.router.ServeHTTP(rw, req)
	return rw
}

func (t *BaseServerTest) requestEmptyBody(c *C) *http.Request {
	return t.requestWithBytes(c, nil)
}

func (t *BaseServerTest) requestWithBytes(c *C, body []byte) *http.Request {
	var httpBody io.Reader
	if body == nil {
		httpBody = nil
	} else {
		httpBody = bytes.NewReader(body)
	}
	return t.requestWithReader(c, httpBody)
}

func (t *BaseServerTest) requestWithReader(c *C, httpBody io.Reader) *http.Request {
	requestURL, err := url.Parse(t.getURL())
	c.Assert(err, IsNil)
	requestURL.RawQuery = t.urlParams.Encode()
	req, err := http.NewRequest(
		t.httpMethod,
		requestURL.String(),
		httpBody)
	c.Assert(err, IsNil)
	if httpBody != nil {
		req.Header.Set("Content-Type", "application/octet-type")
	}
	return req
}

func (t *BaseServerTest) signRequest() {
	SignWithKey(t.req, t.privateKey)
}
