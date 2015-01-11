package api

import (
	"bytes"
	"io"
	"net/http"
	"sort"
	"strconv"

	. "gopkg.in/check.v1"

	"github.com/hoffie/larasync/helpers/bincontainer"
	"github.com/hoffie/larasync/helpers/crypto"
	"github.com/hoffie/larasync/repository"
)

type NIBListTest struct {
	NIBTest
}

var _ = Suite(&NIBListTest{getNIBTest()})

func (t *NIBListTest) SetUpTest(c *C) {
	t.NIBTest.SetUpTest(c)
	t.createRepository(c)
}

func (t *NIBListTest) createNibList(c *C) []*repository.NIB {
	nibs := []*repository.NIB{}
	for i := 0; i < 10; i++ {
		t.setNIBId(
			strconv.FormatInt(int64(i), 10),
		)
		nibs = append(nibs, t.addTestNIB(c))
	}
	return nibs
}

func (t *NIBListTest) TestUnauthorized(c *C) {
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *NIBListTest) TestRepositoryNotExisting(c *C) {
	t.repositoryName = "does-not-exist"
	t.req = t.requestEmptyBody(c)
	t.signRequest()
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *NIBListTest) TestGetCode(c *C) {
	t.signRequest()
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusOK)
}

func (t *NIBListTest) TestGetEmpty(c *C) {
	t.signRequest()
	resp := t.getResponse(t.req)
	decoder := bincontainer.NewDecoder(resp.Body)

	_, err := decoder.ReadChunk()
	c.Assert(err, Equals, io.EOF)
}

func AssertNibSetsEqual(
	c *C,
	storedNibs []*repository.NIB,
	retrievedNibs []*repository.NIB,
) {
	c.Assert(
		len(storedNibs), Equals, len(retrievedNibs),
	)

	nibIds := func(nibs []*repository.NIB) []string {
		nibIds := make([]string, len(nibs))
		for i, nib := range nibs {
			nibIds[i] = nib.ID
		}
		return nibIds
	}

	sortedNibIds := func(nibs []*repository.NIB) []string {
		ids := nibIds(nibs)
		sort.Sort(sort.StringSlice(ids))
		return ids
	}

	storedNibIds := sortedNibIds(storedNibs)
	retrievedNibIds := sortedNibIds(retrievedNibs)

	c.Assert(retrievedNibIds, DeepEquals, storedNibIds)

}

func (t *NIBListTest) nibFromContainer(c *C, data []byte) *repository.NIB {
	reader, err := crypto.NewVerifyingReader(t.pubKey, bytes.NewReader(data))
	c.Assert(err, IsNil)

	nib := repository.NIB{}
	_, err = nib.ReadFrom(reader)
	c.Assert(err, IsNil)

	c.Assert(reader.VerifyAfterRead(), Equals, true)

	return &nib
}

func (t *NIBListTest) getNIBsFromReader(c *C, reader io.Reader) []*repository.NIB {
	decoder := bincontainer.NewDecoder(reader)

	nibs := []*repository.NIB{}
	for {
		data, err := decoder.ReadChunk()

		if err != nil && err != io.EOF {
			c.Error(err)
			break
		} else if err == io.EOF {
			break
		}
		nibs = append(nibs, t.nibFromContainer(c, data))
	}

	return nibs
}

func (t *NIBListTest) TestGetAll(c *C) {
	nibs := t.createNibList(c)
	t.signRequest()
	resp := t.getResponse(t.req)

	AssertNibSetsEqual(c, nibs, t.getNIBsFromReader(c, resp.Body))

}

func (t *NIBListTest) TestGetFromTransactionId(c *C) {
	nibs := t.createNibList(c)

	rep := t.getRepository(c)
	transaction, err := rep.CurrentTransaction()
	c.Assert(err, IsNil)

	fromTransactionID := strconv.FormatInt(transaction.PreviousID, 10)
	t.urlParams.Add("from-transaction-id", fromTransactionID)
	t.req = t.requestEmptyBody(c)
	t.signRequest()
	resp := t.getResponse(t.req)

	c.Assert(resp.Code, Equals, http.StatusOK)

	AssertNibSetsEqual(
		c,
		[]*repository.NIB{nibs[len(nibs)-1]},
		t.getNIBsFromReader(c, resp.Body),
	)

}

// It should have a current transaction id header.
func (t *NIBListTest) TestCurrentTransactionIDHeader(c *C) {
	t.createNibList(c)
	t.signRequest()
	resp := t.getResponse(t.req)

	rep := t.getRepository(c)
	transaction, err := rep.CurrentTransaction()
	c.Assert(err, IsNil)

	c.Assert(
		resp.Header().Get("X-Current-Transaction-Id"),
		Equals,
		strconv.FormatInt(transaction.ID, 10),
	)
}

// It should not have a transaction id if no transaction exists yet in the
// repository.
func (t *NIBListTest) TestNoCurrentTransactionIDHeader(c *C) {
	t.signRequest()
	resp := t.getResponse(t.req)

	c.Assert(
		resp.Header().Get("X-Current-Transaction-Id"),
		Equals,
		"",
	)
}
