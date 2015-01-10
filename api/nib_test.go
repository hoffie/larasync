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

	"github.com/hoffie/larasync/repository"
)

type NIBTest struct {
	BaseTests
	nibID string
}

func generateTestRevision() *repository.Revision {
	return &repository.Revision{
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
			"%s/nibs/%s",
			origGetURL(),
			t.nibID,
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

func (t *NIBTest) getTestNIB() *repository.NIB {
	nib := repository.NIB{}
	nib.ID = t.nibID
	nib.AppendRevision(generateTestRevision())
	return &nib
}

func (t *NIBTest) nibToBytes(nib *repository.NIB) []byte {
	buf := bytes.NewBufferString("")
	nib.WriteTo(buf)
	return buf.Bytes()
}

func (t *NIBTest) getTestNIBBytes() []byte {
	nib := t.getTestNIB()
	return t.nibToBytes(nib)
}

func (t *NIBTest) getTestNIBSignedBytes(c *C) []byte {
	return t.signNIBBytes(c, t.getTestNIBBytes())
}

func (t *NIBTest) addTestNIB(c *C) *repository.NIB {
	return t.addNIB(c, t.getTestNIB())
}

func (t *NIBTest) signNIBBytes(c *C, nibBytes []byte) []byte {
	wr := &bytes.Buffer{}
	signingWriter := repository.NewSigningWriter(t.privateKey, wr)
	_, err := signingWriter.Write(nibBytes)
	c.Assert(err, IsNil)
	err = signingWriter.Finalize()
	c.Assert(err, IsNil)
	return wr.Bytes()
}

func (t *NIBTest) addNIB(c *C, nib *repository.NIB) *repository.NIB {
	repo := t.createRepository(c)
	err := repo.AddNIBContent(
		nib.ID, bytes.NewBuffer(t.signNIBBytes(c, t.nibToBytes(nib))),
	)
	c.Assert(err, IsNil)
	return nib
}

func (t *NIBTest) extractNIB(c *C, resp *httptest.ResponseRecorder) *repository.NIB {
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	reader, err := repository.NewVerifyingReader(
		t.pubKey,
		bytes.NewReader(body),
	)
	c.Assert(err, IsNil)

	nib := repository.NIB{}
	nib.ReadFrom(reader)
	return &nib
}

func (t *NIBTest) verifyNIBSignature(c *C, resp *httptest.ResponseRecorder) bool {
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	reader, err := repository.NewVerifyingReader(
		t.pubKey,
		bytes.NewReader(body),
	)
	c.Assert(err, IsNil)
	_, err = ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	return reader.VerifyAfterRead()
}
