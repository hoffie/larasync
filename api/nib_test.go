package api

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"time"

	. "gopkg.in/check.v1"

	"github.com/hoffie/larasync/helpers/crypto"
	"github.com/hoffie/larasync/repository"
	"github.com/hoffie/larasync/repository/nib"
)

type NIBTest struct {
	BaseTests
	nibID string
}

func generateTestRevision() *nib.Revision {
	return &nib.Revision{
		MetadataID:   "metadataId",
		ContentIDs:   []string{"1", "2", "3"},
		UTCTimestamp: time.Now().UTC().Unix(),
		DeviceID:     "ASDF",
	}
}

func getNIBTest() NIBTest {
	return NIBTest{
		BaseTests: BaseTests{},
	}
}

func (t *NIBTest) SetUpTest(c *C) {
	t.BaseTests.SetUpTest(c)
	origGetURL := t.getURL
	t.getURL = func() string {
		return fmt.Sprintf(
			"%s/nibs",
			origGetURL(),
		)
	}
	t.setNIBId("")
	t.req = t.requestEmptyBody(c)
}

func (t *NIBTest) setNIBId(seed string) string {
	signature := sha512.New()
	signature.Write([]byte(seed))
	t.nibID = hex.EncodeToString(
		signature.Sum(nil),
	)
	return t.nibID
}

func (t *NIBTest) getTestNIB() *nib.NIB {
	n := nib.NIB{}
	n.ID = t.nibID
	n.AppendRevision(generateTestRevision())
	return &n
}

func (t *NIBTest) nibToBytes(n *nib.NIB) []byte {
	buf := bytes.NewBufferString("")
	n.WriteTo(buf)
	return buf.Bytes()
}

func (t *NIBTest) getTestNIBBytes() []byte {
	n := t.getTestNIB()
	return t.nibToBytes(n)
}

func (t *NIBTest) getTestNIBSignedBytes(c *C) []byte {
	return t.signNIBBytes(c, t.getTestNIBBytes())
}

func (t *NIBTest) addTestNIB(c *C) *nib.NIB {
	return t.addNIB(c, t.getTestNIB())
}

func (t *NIBTest) signNIBBytes(c *C, nibBytes []byte) []byte {
	wr := &bytes.Buffer{}
	signingWriter := crypto.NewSigningWriter(t.privateKey, wr)
	_, err := signingWriter.Write(nibBytes)
	c.Assert(err, IsNil)
	err = signingWriter.Finalize()
	c.Assert(err, IsNil)
	return wr.Bytes()
}

func (t *NIBTest) fillNIBContentObjects(c *C, repo *repository.Repository, n *nib.NIB) {
	for _, objectID := range n.AllObjectIDs() {
		if !repo.HasObject(objectID) {
			err := repo.AddObject(objectID, bytes.NewBuffer([]byte("ASDF")))
			c.Assert(err, IsNil)
		}
	}
}

func (t *NIBTest) fillContentOfDefaultNIB(c *C) {
	repo := t.getRepository(c)
	testNib := t.getTestNIB()
	t.fillNIBContentObjects(c, repo, testNib)
}

func (t *NIBTest) addNIB(c *C, n *nib.NIB) *nib.NIB {
	repo := t.createRepository(c)
	t.fillNIBContentObjects(c, repo, n)

	err := repo.AddNIBContent(bytes.NewBuffer(
		t.signNIBBytes(c, t.nibToBytes(n))),
	)
	c.Assert(err, IsNil)
	return n
}

func (t *NIBTest) extractNIB(c *C, resp *httptest.ResponseRecorder) *nib.NIB {
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	reader, err := crypto.NewVerifyingReader(
		t.pubKey,
		bytes.NewReader(body),
	)
	c.Assert(err, IsNil)

	n := nib.NIB{}
	n.ReadFrom(reader)
	return &n
}

func (t *NIBTest) verifyNIBSignature(c *C, resp *httptest.ResponseRecorder) bool {
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	reader, err := crypto.NewVerifyingReader(
		t.pubKey,
		bytes.NewReader(body),
	)
	c.Assert(err, IsNil)
	_, err = ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	return reader.VerifyAfterRead()
}
