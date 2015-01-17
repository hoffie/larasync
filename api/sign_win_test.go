package api

import (
	"bufio"
	"bytes"
	"net/http"

	. "gopkg.in/check.v1"
)

// Tests a real world signature being sent by a windows client.
func (t *SignTests) TestWindowsRequest(c *C) {
	windowsRequest := "PUT /repositories/test HTTP/1.1\r\nHost: 192.168.100.212:14124\r\nUser-Agent: Go 1.1 package http\r\nContent-Length: 58\r\nAuthorization: lara 5549dd45c22957284dd4e0435c521f8e232519af02644ae7fe1f42470ffa6cb2da7344b1f9fea17209766e03266827b644e06a1216d1c5684d22babfa5847c0b\r\nContent-Type: application/json\r\nDate: Sat, 17 Jan 2015 16:51:38 GMT\r\nAccept-Encoding: gzip\r\n\r\n{\"pub_key\":\"83bd+XpfKbkDPx3i6Cg5ClOjWPds0YwktsbdPg6FtAE=\"}"
	windowsRequestBuffer := bytes.NewBuffer([]byte(windowsRequest))
	var err error
	t.req, err = http.ReadRequest(bufio.NewReader(windowsRequestBuffer))
	c.Assert(err, IsNil)

	adminPubkey, err = GetAdminSecretPubkey([]byte("test"))
	if err != nil {
		panic(err)
	}
	sigCheck := validateRequestSig(t.req, adminPubkey)
	c.Assert(sigCheck, Equals, true)
}
