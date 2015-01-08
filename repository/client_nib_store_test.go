package repository

import (
	"bytes"
	//	"crypto/sha256"
	//	"encoding/hex"
	//	"io"
	"io/ioutil"
	//	"os"
	//	"path"
	"path/filepath"

	"github.com/agl/ed25519"

	. "gopkg.in/check.v1"
)

type ClientNIBStoreTest struct {
	dir                string
	repository         *Repository
	nibStore           *ClientNIBStore
	storage            ContentStorage
	transactionManager *TransactionManager
}

var _ = Suite(&ClientNIBStoreTest{})

func (t *ClientNIBStoreTest) SetUpTest(c *C) {
	t.dir = c.MkDir()
	t.repository = New(filepath.Join(t.dir, "repo"))
	t.repository.Create()

	pubKey, privKey, err := ed25519.GenerateKey(
		bytes.NewBufferString("just some deterministic 'random' bytes"))
	c.Assert(err, IsNil)

	err = t.repository.SetSigningPrivkey(*privKey)
	c.Assert(err, IsNil)
	signingPubKey, err := t.repository.GetSigningPubkey()
	c.Assert(*pubKey, DeepEquals, signingPubKey)

	fileStorage := FileContentStorage{
		StoragePath: filepath.Join(t.dir, "nibs"),
	}
	err = fileStorage.CreateDir()
	c.Assert(err, IsNil)

	storage := ContentStorage(fileStorage)
	t.storage = storage

	transactionStorage := FileContentStorage{
		StoragePath: filepath.Join(t.dir, "transactions"),
	}
	err = transactionStorage.CreateDir()
	c.Assert(err, IsNil)

	t.transactionManager = newTransactionManager(transactionStorage)
	t.nibStore = newClientNIBStore(
		&t.storage,
		t.repository,
		t.transactionManager,
	)
}

func (t *ClientNIBStoreTest) getTestNIB() *NIB {
	n := &NIB{}
	n.AppendRevision(&Revision{})
	n.AppendRevision(&Revision{})
	return n
}

func (t *ClientNIBStoreTest) addTestNIB(c *C) *NIB {
	n := t.getTestNIB()
	err := t.nibStore.Add(n)
	c.Assert(err, IsNil)
	return n
}

// It should be able to add a NIB
func (t *ClientNIBStoreTest) TestNibAddition(c *C) {
	testNib := t.addTestNIB(c)
	c.Assert(testNib.UUID, Not(Equals), "")
}

// It should create a transaction with the NIBs UUID on
// addition.
func (t *ClientNIBStoreTest) TestTransactionAddition(c *C) {
	testNib := t.addTestNIB(c)
	transaction, err := t.transactionManager.Get(testNib.UUID)
	c.Assert(err, IsNil)
	c.Assert(transaction.UUID, Equals, testNib.UUID)
}

func (t *ClientNIBStoreTest) TestNibGet(c *C) {
	testNib := t.addTestNIB(c)
	nib, err := t.nibStore.Get(testNib.UUID)
	c.Assert(err, IsNil)
	c.Assert(nib.UUID, Equals, testNib.UUID)
}

func (t *ClientNIBStoreTest) TestNibGetSignatureMangled(c *C) {
	testNib := t.addTestNIB(c)
	reader, err := t.storage.Get(testNib.UUID)
	data, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	data[len(data)-1] = 50
	t.storage.Set(testNib.UUID, bytes.NewReader(data))
	_, err = t.nibStore.Get(testNib.UUID)
	c.Assert(err, NotNil)
}

func (t *ClientNIBStoreTest) TestNibGetContentMangled(c *C) {
	testNib := t.addTestNIB(c)
	reader, err := t.storage.Get(testNib.UUID)
	data, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	data[0] = 50
	t.storage.Set(testNib.UUID, bytes.NewReader(data))
	_, err = t.nibStore.Get(testNib.UUID)
	c.Assert(err, NotNil)
}

func (t *ClientNIBStoreTest) TestNibExistsPositive(c *C) {
	testNib := t.addTestNIB(c)
	c.Assert(t.nibStore.Exists(testNib.UUID), Equals, true)
}

func (t *ClientNIBStoreTest) TestNibExistsNegative(c *C) {
	c.Assert(t.nibStore.Exists("Does not exist"), Equals, false)
}

func (t *ClientNIBStoreTest) TestNibVerificationSignatureError(c *C) {
	testNib := t.addTestNIB(c)
	reader, err := t.storage.Get(testNib.UUID)
	data, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	data[len(data)-1] = 50

	reader = bytes.NewReader(data)

	c.Assert(
		t.nibStore.VerifyContent(reader), Equals, ErrSignatureVerification,
	)
}

func (t *ClientNIBStoreTest) TestNibVerificationMarshallingError(c *C) {
	testNib := t.addTestNIB(c)
	reader, err := t.storage.Get(testNib.UUID)
	data, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	data[0] = 50

	reader = bytes.NewReader(data)

	c.Assert(
		t.nibStore.VerifyContent(reader), Equals, ErrUnMarshalling,
	)
}

func (t *ClientNIBStoreTest) TestNibVerification(c *C) {
	testNib := t.addTestNIB(c)
	reader, err := t.storage.Get(testNib.UUID)
	c.Assert(err, IsNil)

	c.Assert(
		t.nibStore.VerifyContent(reader), IsNil)
}
