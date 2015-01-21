package client

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"io/ioutil"

	. "gopkg.in/check.v1"
)

type ObjectsClientTest struct {
	BaseTest
	data      []byte
	objectKey string
}

var _ = Suite(&ObjectsClientTest{
BaseTest: newBaseTest(),
})

func (t *ObjectsClientTest) SetUpTest(c *C) {
	t.BaseTest.SetUpTest(c)
	t.data = []byte("This is a testfile.")

	h := sha512.New()
	h.Write(t.data)
	t.objectKey = hex.EncodeToString(h.Sum(nil))
	t.createRepository(c)
}

func (t *ObjectsClientTest) TestNotExisting(c *C) {
	_, err := t.client.GetObject(t.objectKey)
	c.Assert(err, NotNil)
}

func (t *ObjectsClientTest) TestExisting(c *C) {
	repository := t.getRepository(c)
	err := repository.AddObject(t.objectKey, bytes.NewReader(t.data))
	c.Assert(err, IsNil)

	reader, err := t.client.GetObject(t.objectKey)
	c.Assert(err, IsNil)

	data, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	c.Assert(data, DeepEquals, t.data)
}

func (t *ObjectsClientTest) TestConnError(c *C) {
	t.server.Close()
	_, err := t.client.GetObject(t.objectKey)
	c.Assert(err, NotNil)
}

func (t *ObjectsClientTest) TestCreate(c *C) {
	err := t.client.PutObject(t.objectKey, bytes.NewReader(t.data))
	c.Assert(err, IsNil)

	repository := t.getRepository(c)
	reader, err := repository.GetObjectData(t.objectKey)
	c.Assert(err, IsNil)
	defer reader.Close()
	d, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	c.Assert(d, DeepEquals, t.data)
}

func (t *ObjectsClientTest) TestCreateConnError(c *C) {
	t.server.Close()
	err := t.client.PutObject(t.objectKey, bytes.NewReader(t.data))
	c.Assert(err, NotNil)
}
