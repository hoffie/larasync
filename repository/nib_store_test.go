package repository

import (
	"bytes"
	"fmt"
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

	err = t.repository.keys.SetSigningPrivateKey(*privKey)
	c.Assert(err, IsNil)
	signingPubKey, err := t.repository.keys.SigningPublicKey()
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

	t.transactionManager = newTransactionManager(
		transactionStorage,
		t.repository.GetManagementDir())
	t.nibStore = newNIBStore(
		t.storage,
		t.repository.keys,
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

	c.Assert(
		t.nibStore.VerifyContent(data), Equals, ErrSignatureVerification,
	)
}

func (t *NIBStoreTest) TestNibVerificationMarshallingError(c *C) {
	testNib := t.addTestNIB(c)
	reader, err := t.storage.Get(testNib.ID)
	data, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	data[0] = 50

	c.Assert(
		t.nibStore.VerifyContent(data), Equals, ErrUnMarshalling,
	)
}

func (t *NIBStoreTest) TestNibVerification(c *C) {
	testNib := t.addTestNIB(c)
	reader, err := t.storage.Get(testNib.ID)
	c.Assert(err, IsNil)

	data, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)

	c.Assert(
		t.nibStore.VerifyContent(data), IsNil)
}

// It should return all added bytes
func (t *NIBStoreTest) TestGetAllBytes(c *C) {
	for i := 0; i < 100; i++ {
		nib := t.getTestNIB()
		nib.ID = fmt.Sprintf("test%d", i)
		t.nibStore.Add(nib)
	}
	found := 0

	channel, err := t.nibStore.GetAllBytes()
	c.Assert(err, IsNil)

	for _ = range channel {
		found += 1
	}

	c.Assert(found, Equals, 100)
}

// It should Return all nib bytes
func (t *NIBStoreTest) TestGetAll(c *C) {
	for i := 0; i < 100; i++ {
		nib := t.getTestNIB()
		nib.ID = fmt.Sprintf("test%d", i)
		t.nibStore.Add(nib)
	}
	seen := []string{}

	channel, err := t.nibStore.GetAll()
	c.Assert(err, IsNil)

	for nib := range channel {
		if nib == nil {
			c.Error("error from channel")
			return
		}
		for _, existingNib := range seen {
			if existingNib == nib.ID {
				c.Error("Double nib found.")
			}
		}
		seen = append(seen, nib.ID)
	}

	c.Assert(len(seen), Equals, 100)

}

func (t *NIBStoreTest) TestGetFrom(c *C) {
	for i := 0; i < 100; i++ {
		nib := t.getTestNIB()
		nib.ID = fmt.Sprintf("test%d", i)
		t.nibStore.Add(nib)
	}

	transaction, err := t.transactionManager.CurrentTransaction()
	c.Assert(err, IsNil)
	for i := 0; i < 5; i++ {
		transaction, err = t.transactionManager.Get(transaction.PreviousID)
		c.Assert(err, IsNil)
	}

	expectedIds := []string{}
	for i := 95; i < 100; i++ {
		expectedIds = append(expectedIds, fmt.Sprintf("test%d", i))
	}
	foundIds := []string{}
	channel, err := t.nibStore.GetFrom(transaction.ID)

	for nib := range channel {
		foundIds = append(foundIds, nib.ID)
	}

	c.Assert(expectedIds, DeepEquals, foundIds)
}

func (t *NIBStoreTest) TestGetBytesFrom(c *C) {
	for i := 0; i < 100; i++ {
		nib := t.getTestNIB()
		nib.ID = fmt.Sprintf("test%d", i)
		t.nibStore.Add(nib)
	}

	transaction, err := t.transactionManager.CurrentTransaction()
	c.Assert(err, IsNil)
	for i := 0; i < 5; i++ {
		transaction, err = t.transactionManager.Get(transaction.PreviousID)
		c.Assert(err, IsNil)
	}

	channel, err := t.nibStore.GetBytesFrom(transaction.ID)

	found := 0
	for _ = range channel {
		found += 1
	}

	c.Assert(found, Equals, 5)
}
