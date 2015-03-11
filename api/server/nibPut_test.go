package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/helpers"
	"github.com/hoffie/larasync/repository/nib"

	. "gopkg.in/check.v1"
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

	n, err := r.GetNIB(t.nibID)
	c.Assert(err, IsNil)
	c.Assert(n.ID, Equals, t.nibID)
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

func (t *NIBPutTest) changeNIBForPut(c *C, n *nib.NIB) *nib.NIB {
	revision := generateTestRevision()
	revision.MetadataID = "other-metadata"
	revision.DeviceID = "other-id"

	n.AppendRevision(revision)
	return n
}

func (t *NIBPutTest) requestWithNib(c *C, n *nib.NIB) *http.Request {
	signedData := t.signNIBBytes(
		c,
		t.nibToBytes(n),
	)
	return t.requestWithBytes(c, signedData)
}

func (t *NIBPutTest) TestPutUpdate(c *C) {
	n := t.addTestNIB(c)
	n = t.changeNIBForPut(c, n)
	repo := t.getRepository(c)
	t.fillNIBContentObjects(c, repo, n)
	t.req = t.requestWithNib(c, n)
	t.signRequest()

	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusOK)
}

func (t *NIBPutTest) TestPutChanged(c *C) {
	n := t.normalPut(c)
	repo := t.getRepository(c)
	repoNib, err := repo.GetNIB(n.ID)
	c.Assert(err, IsNil)

	revisions := repoNib.Revisions
	c.Assert(len(revisions), Equals, 2)

	c.Assert(revisions[1].DeviceID, Equals, "other-id")
}

func (t *NIBPutTest) TestPutConflict(c *C) {
	n := t.normalPut(c)
	latestRev, err := n.LatestRevision()
	c.Assert(err, IsNil)
	latestRev.ContentIDs = []string{"changed"}
	repo := t.getRepository(c)
	t.fillNIBContentObjects(c, repo, n)
	t.req = t.requestWithNib(c, n)
	t.signRequest()

	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusConflict)

}

func (t *NIBPutTest) normalPut(c *C) *nib.NIB {
	n := t.addTestNIB(c)
	n = t.changeNIBForPut(c, n)
	repo := t.getRepository(c)
	t.fillNIBContentObjects(c, repo, n)

	t.req = t.requestWithNib(c, n)
	t.signRequest()

	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusOK)
	return n
}

func (t *NIBPutTest) TestPutReferencedContentMissing(c *C) {
	n := t.addTestNIB(c)
	n = t.changeNIBForPut(c, n)

	t.req = t.requestWithNib(c, n)
	t.signRequest()

	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusPreconditionFailed)
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)

	jsonError := &api.ContentIDsJSONError{}
	err = json.Unmarshal(data, jsonError)
	c.Assert(err, IsNil)
	for _, objectID := range n.AllObjectIDs() {
		helpers.SliceContainsString(jsonError.MissingContentIDs, objectID)
	}
}
