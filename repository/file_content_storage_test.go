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

type FileContentStorageTests struct {
	dir     string
	storage *FileContentStorage
	data    []byte
}

var _ = Suite(&FileContentStorageTests{})

func (t *FileContentStorageTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	t.storage = &FileContentStorage{StoragePath: t.dir}
	t.data = []byte("This is a test blob storage file input.")
}

func (t *FileContentStorageTests) blobID() string {
	blobIDBytes := sha256.New().Sum(t.data)
	return hex.EncodeToString(blobIDBytes[:])
}

func (t *FileContentStorageTests) blobPath() string {
	return path.Join(t.dir, t.blobID())
}

func (t *FileContentStorageTests) testReader() io.Reader {
	return bytes.NewReader(t.data)
}

func (t *FileContentStorageTests) setData() error {
	return t.storage.Set(t.blobID(), t.testReader())
}

func (t *FileContentStorageTests) TestSet(c *C) {
	err := t.setData()
	c.Assert(err, IsNil)
	_, err = os.Stat(t.blobPath())
	c.Assert(err, IsNil)
}

func (t *FileContentStorageTests) TestSetInputData(c *C) {
	t.setData()
	file, _ := os.Open(t.blobPath())
	fileData, _ := ioutil.ReadAll(file)
	c.Assert(fileData[:], DeepEquals, t.data[:])
}

func (t *FileContentStorageTests) TestExistsNegative(c *C) {
	c.Assert(t.storage.Exists(t.blobID()), Equals, false)
}

func (t *FileContentStorageTests) TestExistsPositive(c *C) {
	t.setData()
	c.Assert(t.storage.Exists(t.blobID()), Equals, true)
}

func (t *FileContentStorageTests) TestGet(c *C) {
	t.storage.Set(t.blobID(), t.testReader())
	_, err := t.storage.Get(t.blobID())
	c.Assert(err, IsNil)
}

func (t *FileContentStorageTests) TestGetData(c *C) {
	t.setData()
	file, _ := t.storage.Get(t.blobID())
	fileData, _ := ioutil.ReadAll(file)
	c.Assert(fileData[:], DeepEquals, t.data)
}

func (t *FileContentStorageTests) TestGetError(c *C) {
	_, err := t.storage.Get(t.blobID())
	c.Assert(err, NotNil)
}

func (t *FileContentStorageTests) TestSetError(c *C) {
	os.RemoveAll(t.dir)

	err := t.storage.Set(t.blobID(),
		t.testReader())
	c.Assert(err, NotNil)
}

func (t *FileContentStorageTests) TestDelete(c *C) {
	t.setData()
	err := t.storage.Delete(t.blobID())
	c.Assert(err, IsNil)
	c.Assert(t.storage.Exists(t.blobID()), Equals, false)
}

func (t *FileContentStorageTests) TestDeleteError(c *C) {
	err := t.storage.Delete(t.blobID())
	c.Assert(err, NotNil)
}
