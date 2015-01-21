package server

import (
	"bytes"
	"fmt"
	"io/ioutil"

	. "gopkg.in/check.v1"
)

type BlobTests struct {
	BaseTests
	blobID   string
	blobData []byte
}

func createBlobTests() BlobTests {
	return BlobTests{
		BaseTests: newBaseTest(),
	}
}

func (t *BlobTests) SetUpTest(c *C) {
	t.BaseTests.SetUpTest(c)
	t.blobID = "1234567890"
	t.blobData = []byte("This is testdata")
	baseURLGet := t.getURL
	t.getURL = func() string {
		return fmt.Sprintf(
			"%s/blobs/%s",
			baseURLGet(),
			t.blobID,
		)
	}
	t.req = t.requestWithBytes(c, nil)
}

func (t *BlobTests) createBlob(c *C) {
	t.createBlobWithData(c, t.blobData)
}

func (t *BlobTests) createBlobWithData(c *C, data []byte) {
	if !t.rm.Exists(t.repositoryName) {
		t.createRepository(c)
	}

	repository, err := t.rm.Open(t.repositoryName)
	c.Assert(err, IsNil)
	err = repository.AddObject(t.blobID, bytes.NewReader(data))
	c.Assert(err, IsNil)
}

func (t *BlobTests) expectStoredBlobData(c *C) {
	t.expectStoredData(c, t.blobData)
}

func (t *BlobTests) expectStoredData(c *C, expectedData []byte) {
	repository, err := t.rm.Open(t.repositoryName)
	c.Assert(err, IsNil)
	reader, err := repository.GetObjectData(t.blobID)
	c.Assert(err, IsNil)
	retrievedData, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	err = reader.Close()
	c.Assert(err, IsNil)
	c.Assert(retrievedData, DeepEquals, expectedData)
}
