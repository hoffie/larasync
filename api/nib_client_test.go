package api

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type NIBClientTest struct {
	BaseClientTest
	data []byte
}

var _ = Suite(&NIBClientTest{
	BaseClientTest: newBaseClientTest(),
})

func (t *NIBClientTest) SetUpTest(c *C) {
	t.BaseClientTest.SetUpTest(c)

	t.createRepository(c)
	t.data = []byte("This is testdata")
}

func (t *NIBClientTest) AddTestData(c *C) {
	repository := t.getClientRepository(c)
	path := filepath.Join(t.repositoryPath(c), "test.txt")
	err := ioutil.WriteFile(path, t.data, 0600)
	c.Assert(err, IsNil)

	err = repository.AddItem(path)
	c.Assert(err, IsNil)
}

func (t *NIBClientTest) TestGet(c *C) {
	t.AddTestData(c)
	channel, err := t.client.GetNIBs()
	c.Assert(err, IsNil)
	i := 0
	for _ = range channel {
		i++
	}
	c.Assert(i, Equals, 1)
}

func (t *NIBClientTest) TestConnError(c *C) {
	t.server.Close()
	_, err := t.client.GetNIBs()
	c.Assert(err, NotNil)
}

func (t *NIBClientTest) prepareForNIBAddition(c *C) (string, []byte) {
	repository := t.getClientRepository(c)
	t.AddTestData(c)
	channel, err := repository.GetAllNibs()
	c.Assert(err, IsNil)
	nib := <-channel

	reader, err := repository.GetNIBReader(nib.ID)
	c.Assert(err, IsNil)
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	return nib.ID, data

}

func (t *NIBClientTest) TestAdd(c *C) {
	ID, data := t.prepareForNIBAddition(c)

	err := t.client.PutNIB(ID, bytes.NewReader(data))
	c.Assert(err, IsNil)
}

func (t *NIBClientTest) TestAddConnError(c *C) {
	ID, data := t.prepareForNIBAddition(c)
	t.server.Close()

	err := t.client.PutNIB(ID, bytes.NewReader(data))
	c.Assert(err, NotNil)
}
