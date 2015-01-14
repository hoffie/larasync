package api

import (
	"net/http"

	. "gopkg.in/check.v1"

	"github.com/hoffie/larasync/repository"
)

type NIBPutTest struct {
	NIBItemTest
}

var _ = Suite(&NIBPutTest{getNIBItemTest()})

func (t *NIBPutTest) SetUpTest(c *C) {
	t.NIBItemTest.SetUpTest(c)
	t.httpMethod = "PUT"
	t.req = t.requestWithBytes(c, t.signNIBBytes(c, t.getTestNIBBytes()))
	t.createRepository(c)
}

func (t *NIBPutTest) TestUnauthorized(c *C) {
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *NIBPutTest) TestRepositoryNotExisting(c *C) {
	t.repositoryName = "does-not-exist"
	t.req = t.requestEmptyBody(c)
	t.signRequest()
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *NIBPutTest) TestPutMalformedSignature(c *C) {
	data := t.getTestNIBBytes()
	data = t.signNIBBytes(c, data)

	// destroy signature:
	for x := 0; x < SignatureSize; x++ {
		data[len(data)-1-x] = 0
	}
	t.req = t.requestWithBytes(c, data)
	t.signRequest()

	resp := t.getResponse(t.req)

	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
	repo := t.getRepository(c)
	c.Assert(repo.HasNIB(t.nibID), Equals, false)
}

func (t *NIBPutTest) TestPutMalformedData(c *C) {
	data := t.getTestNIBBytes()
	data[0] = 0
	data = t.signNIBBytes(c, data)
	t.req = t.requestWithBytes(c, data)
	t.signRequest()

	resp := t.getResponse(t.req)

	c.Assert(resp.Code, Equals, http.StatusBadRequest)
	repo := t.getRepository(c)
	c.Assert(repo.HasNIB(t.nibID), Equals, false)
}

func (t *NIBPutTest) TestPutNew(c *C) {
	t.fillContentOfDefaultNIB(c)
	t.signRequest()
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusCreated)
}

func (t *NIBPutTest) TestPutNewStore(c *C) {
	r := t.getRepository(c)
	t.fillContentOfDefaultNIB(c)

	t.signRequest()
	t.getResponse(t.req)

	nib, err := r.GetNIB(t.nibID)
	c.Assert(err, IsNil)
	c.Assert(nib.ID, Equals, t.nibID)
}

func (t *NIBPutTest) TestPutNewPrecondition(c *C) {
	t.addTestNIB(c)
	header := t.req.Header
	rep := t.getRepository(c)
	transaction, err := rep.CurrentTransaction()
	c.Assert(err, IsNil)
	header.Set("If-Match", transaction.IDString())
	t.signRequest()

	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusOK)
}

func (t *NIBPutTest) TestPutNewPreconditionFailed(c *C) {
	t.addTestNIB(c)
	header := t.req.Header
	rep := t.getRepository(c)
	transaction, err := rep.CurrentTransaction()
	c.Assert(err, IsNil)
	header.Set("If-Match", transaction.PreviousIDString())
	t.signRequest()

	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusPreconditionFailed)
}

func (t *NIBPutTest) changeNIBForPut(c *C, nib *repository.NIB) *repository.NIB {
	revision := generateTestRevision()
	revision.MetadataID = "other-metadata"
	revision.DeviceID = "other-id"

	nib.AppendRevision(revision)
	return nib
}

func (t *NIBPutTest) requestWithNib(c *C, nib *repository.NIB) *http.Request {
	signedData := t.signNIBBytes(
		c,
		t.nibToBytes(nib),
	)
	return t.requestWithBytes(c, signedData)
}

func (t *NIBPutTest) TestPutUpdate(c *C) {
	nib := t.addTestNIB(c)
	nib = t.changeNIBForPut(c, nib)
	repo := t.getRepository(c)
	t.fillNIBContentObjects(c, repo, nib)
	t.req = t.requestWithNib(c, nib)
	t.signRequest()

	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusOK)
}

func (t *NIBPutTest) TestPutChanged(c *C) {
	nib := t.addTestNIB(c)
	nib = t.changeNIBForPut(c, nib)
	repo := t.getRepository(c)
	t.fillNIBContentObjects(c, repo, nib)

	t.req = t.requestWithNib(c, nib)
	t.signRequest()

	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusOK)
	repoNib, err := repo.GetNIB(nib.ID)
	c.Assert(err, IsNil)

	revisions := repoNib.Revisions
	c.Assert(len(revisions), Equals, 2)

	c.Assert(revisions[1].DeviceID, Equals, "other-id")
}

func (t *NIBPutTest) TestPutReferencedContentMissing(c *C) {
	t.signRequest()
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusPreconditionFailed)
}
