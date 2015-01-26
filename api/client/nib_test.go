package client

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type NIBClientTest struct {
	BaseTest
	data []byte
}

var _ = Suite(&NIBClientTest{
	BaseTest: newBaseTest(),
})

func (t *NIBClientTest) SetUpTest(c *C) {
	t.BaseTest.SetUpTest(c)

	t.createRepository(c)
	t.data = []byte("This is testdata")
}

func (t *NIBClientTest) AddTestData(c *C) {
	t.AddDataWith(c, t.data)
}

func (t *NIBClientTest) AddDataWith(c *C, content []byte) {
	repository := t.getClientRepository(c)
	path := filepath.Join(t.repositoryPath(c), "test.txt")
	err := ioutil.WriteFile(path, content, 0600)
	c.Assert(err, IsNil)

	err = repository.AddItem(path)
	c.Assert(err, IsNil)
}

func (t *NIBClientTest) getGETResponse(c *C) *NIBGetResponse {
	t.AddTestData(c)

	response, err := t.client.GetNIBs()
	c.Assert(err, IsNil)
	return response
}

func (t *NIBClientTest) TestGet(c *C) {
	response := t.getGETResponse(c)
	i := 0
	for _ = range response.NIBData {
		i++
	}
	c.Assert(i, Equals, 1)
}

func (t *NIBClientTest) TestGetTransactionIDResponse(c *C) {
	repository := t.getClientRepository(c)
	transaction, err := repository.CurrentTransaction()
	c.Assert(err, IsNil)
	response := t.getGETResponse(c)
	c.Assert(response.ServerTransactionID, Equals, transaction.ID)
}

func (t *NIBClientTest) getFromTransactionIDResponse(c *C) *NIBGetResponse {
	t.AddTestData(c)
	t.AddDataWith(c, []byte("Hello world"))
	repository := t.getClientRepository(c)
	transaction, err := repository.CurrentTransaction()
	c.Assert(err, IsNil)
	response, err := t.client.GetNIBsFromTransactionID(transaction.ID - 1)
	c.Assert(err, IsNil)
	return response
}

func (t *NIBClientTest) TestGetFromTransactionID(c *C) {
	response := t.getFromTransactionIDResponse(c)
	i := 0
	for _ = range response.NIBData {
		i++
	}
	c.Assert(i, Equals, 1)
}

func (t *NIBClientTest) TestGetFromTransactionIDResponse(c *C) {
	response := t.getFromTransactionIDResponse(c)
	repository := t.getRepository(c)
	transaction, err := repository.CurrentTransaction()
	c.Assert(err, IsNil)
	c.Assert(response.ServerTransactionID, Equals, transaction.ID)
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
