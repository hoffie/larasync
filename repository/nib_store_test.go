package repository

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	"github.com/agl/ed25519"

	. "gopkg.in/check.v1"
)

type NIBStoreTest struct {
	dir                string
	repository         *Repository
	nibStore           *NIBStore
	storage            ContentStorage
	transactionManager *TransactionManager
}

var _ = Suite(&NIBStoreTest{})

func (t *NIBStoreTest) SetUpTest(c *C) {
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

	t.storage = fileStorage

	transactionStorage := FileContentStorage{
		StoragePath: filepath.Join(t.dir, "transactions"),
	}
	err = transactionStorage.CreateDir()
	c.Assert(err, IsNil)

	t.transactionManager = newTransactionManager(transactionStorage)
	t.nibStore = newNIBStore(
		t.storage,
		t.repository,
		t.transactionManager,
	)
}

func (t *NIBStoreTest) getTestNIB() *NIB {
	n := &NIB{ID: "test"}
	n.AppendRevision(&Revision{})
	n.AppendRevision(&Revision{})
	return n
}

func (t *NIBStoreTest) addTestNIB(c *C) *NIB {
	n := t.getTestNIB()
	err := t.nibStore.Add(n)
	c.Assert(err, IsNil)
	return n
}

// It should be able to add a NIB
func (t *NIBStoreTest) TestNibAddition(c *C) {
	testNib := t.addTestNIB(c)
	c.Assert(testNib.ID, Not(Equals), "")
}

// It should create a transaction with the NIBs ID on
// addition.
func (t *NIBStoreTest) TestTransactionAddition(c *C) {
	testNib := t.addTestNIB(c)
	transaction, err := t.transactionManager.CurrentTransaction()
	c.Assert(err, IsNil)
	c.Assert(transaction.NIBIDs, DeepEquals, []string{testNib.ID})
}

func (t *NIBStoreTest) TestNibGet(c *C) {
	testNib := t.addTestNIB(c)
	nib, err := t.nibStore.Get(testNib.ID)
	c.Assert(err, IsNil)
	c.Assert(nib.ID, Equals, testNib.ID)
}

func (t *NIBStoreTest) TestNibGetSignatureMangled(c *C) {
	testNib := t.addTestNIB(c)
	reader, err := t.storage.Get(testNib.ID)
	data, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	data[len(data)-1] = 50
	t.storage.Set(testNib.ID, bytes.NewReader(data))
	_, err = t.nibStore.Get(testNib.ID)
	c.Assert(err, NotNil)
}

func (t *NIBStoreTest) TestNibGetContentMangled(c *C) {
	testNib := t.addTestNIB(c)
	reader, err := t.storage.Get(testNib.ID)
	data, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	data[0] = 50
	t.storage.Set(testNib.ID, bytes.NewReader(data))
	_, err = t.nibStore.Get(testNib.ID)
	c.Assert(err, NotNil)
}

func (t *NIBStoreTest) TestNibExistsPositive(c *C) {
	testNib := t.addTestNIB(c)
	c.Assert(t.nibStore.Exists(testNib.ID), Equals, true)
}

func (t *NIBStoreTest) TestNibExistsNegative(c *C) {
	c.Assert(t.nibStore.Exists("Does not exist"), Equals, false)
}

func (t *NIBStoreTest) TestNibVerificationSignatureError(c *C) {
	testNib := t.addTestNIB(c)
	reader, err := t.storage.Get(testNib.ID)
	data, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	data[len(data)-1] = 50

	reader = bytes.NewReader(data)

	c.Assert(
		t.nibStore.VerifyContent(reader), Equals, ErrSignatureVerification,
	)
}

func (t *NIBStoreTest) TestNibVerificationMarshallingError(c *C) {
	testNib := t.addTestNIB(c)
	reader, err := t.storage.Get(testNib.ID)
	data, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	data[0] = 50

	reader = bytes.NewReader(data)

	c.Assert(
		t.nibStore.VerifyContent(reader), Equals, ErrUnMarshalling,
	)
}

func (t *NIBStoreTest) TestNibVerification(c *C) {
	testNib := t.addTestNIB(c)
	reader, err := t.storage.Get(testNib.ID)
	c.Assert(err, IsNil)

	c.Assert(
		t.nibStore.VerifyContent(reader), IsNil)
}
