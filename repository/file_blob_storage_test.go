package repository

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path"

	. "gopkg.in/check.v1"
)

type FileBlobStorageTests struct {
	dir     string
	storage *FileBlobStorage
	data    []byte
}

var _ = Suite(&FileBlobStorageTests{})

func (t *FileBlobStorageTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	t.storage = &FileBlobStorage{StoragePath: t.dir}
	t.data = []byte("This is a test blob storage file input.")
}

func (t *FileBlobStorageTests) blobId() string {
	blobIdBytes := sha256.New().Sum(t.data)
	return hex.EncodeToString(blobIdBytes[:])
}

func (t *FileBlobStorageTests) blobPath() string {
	return path.Join(t.dir, t.blobId())
}

func (t *FileBlobStorageTests) testReader() io.Reader {
	return bytes.NewReader(t.data)
}

func (t *FileBlobStorageTests) setData() error {
	return t.storage.Set(t.blobId(), t.testReader())
}

func (t *FileBlobStorageTests) TestSet(c *C) {
	err := t.setData()
	c.Assert(err, IsNil)
	_, err = os.Stat(t.blobPath())
	c.Assert(err, IsNil)
}

func (t *FileBlobStorageTests) TestSetInputData(c *C) {
	t.setData()
	file, _ := os.Open(t.blobPath())
	fileData, _ := ioutil.ReadAll(file)
	c.Assert(fileData[:], DeepEquals, t.data[:])
}

func (t *FileBlobStorageTests) TestExistsNegative(c *C) {
	c.Assert(t.storage.Exists(t.blobId()), Equals, false)
}

func (t *FileBlobStorageTests) TestExistsPositive(c *C) {
	t.setData()
	c.Assert(t.storage.Exists(t.blobId()), Equals, true)
}

func (t *FileBlobStorageTests) TestGet(c *C) {
	t.storage.Set(t.blobId(), t.testReader())
	_, err := t.storage.Get(t.blobId())
	c.Assert(err, IsNil)
}

func (t *FileBlobStorageTests) TestGetData(c *C) {
	t.setData()
	file, _ := t.storage.Get(t.blobId())
	fileData, _ := ioutil.ReadAll(file)
	c.Assert(fileData[:], DeepEquals, t.data)
}

func (t *FileBlobStorageTests) TestGetError(c *C) {
	_, err := t.storage.Get(t.blobId())
	c.Assert(err, NotNil)
}

func (t *FileBlobStorageTests) TestSetError(c *C) {
	os.RemoveAll(t.dir)

	err := t.storage.Set(t.blobId(),
		t.testReader())
	c.Assert(err, NotNil)
}