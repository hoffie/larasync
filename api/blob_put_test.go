package api

import (
	"net/http"

	. "gopkg.in/check.v1"
)

type BlobPutTests struct {
	BlobTests
}

var _ = Suite(&BlobPutTests{BlobTests{}})

func (t *BlobPutTests) SetUpTest(c *C) {
	t.BlobTests.SetUpTest(c)
	t.httpMethod = "PUT"
	t.req = t.requestWithBytes(c, t.blobData)
}

func (t *BlobPutTests) TestRepoAccessUnauthorized(c *C) {
	t.createRepository(c)
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

// Should return ok if the blob does not exist yet.
func (t *BlobPutTests) TestBlobCreateStatus(c *C) {
	t.createRepository(c)
	t.signRequest()
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusOK)
}

// Should create a blob with the given content if the blob does not
// exist yet.
func (t *BlobPutTests) TestBlobCreateData(c *C) {
	t.createRepository(c)
	t.signRequest()
	t.getResponse(t.req)
	t.expectStoredBlobData(c)
}

func (t *BlobPutTests) TestBlobOverwrite(c *C) {
	t.createRepository(c)
	t.createBlobWithData(c, []byte("Other test data"))
	t.signRequest()
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusOK)
	t.expectStoredBlobData(c)
}

func (t *BlobPutTests) TestBlobPutLocationHeader(c *C) {
	t.createRepository(c)
	t.createBlob(c)
	t.signRequest()
	resp := t.getResponse(t.req)
	location := resp.Header().Get("Location")
	c.Assert(location, Equals, t.req.URL.String())
}
