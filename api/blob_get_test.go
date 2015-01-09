package api

import (
	"io/ioutil"
	"net/http"

	. "gopkg.in/check.v1"
)

type BlobGetTests struct {
	BlobTests
}

var _ = Suite(
	&BlobGetTests{
		createBlobTests(),
	},
)

func (t *BlobGetTests) TestRepoAccessUnauthorized(c *C) {
	t.createRepository(c)
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

// Should return unauthorized if repository does not exist.
func (t *BlobGetTests) TestRepoAccessNotFound(c *C) {
	t.signRequest()
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

// Should return not found if blobID does not exist.
func (t *BlobGetTests) TestBlobNotFound(c *C) {
	t.createRepository(c)
	t.signRequest()
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusNotFound)
}

func (t *BlobGetTests) TestBlobFound(c *C) {
	t.createRepository(c)
	t.createBlob(c)
	t.signRequest()
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusOK)
}

func (t *BlobGetTests) TestBlobGetData(c *C) {
	t.createRepository(c)
	t.createBlob(c)
	t.signRequest()
	resp := t.getResponse(t.req)
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(data, DeepEquals, t.blobData)
}
