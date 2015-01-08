package repository

import (
	"os"

	. "gopkg.in/check.v1"
)

type TransactionContainerManagerTest struct {
	dir string
	tcm *TransactionContainerManager
}

var _ = Suite(&TransactionContainerManagerTest{})

func (t *TransactionContainerManagerTest) SetUpTest(c *C) {
	t.dir = c.MkDir()
	storage := &FileContentStorage{StoragePath: t.dir}

	t.tcm = newTransactionContainerManager(storage, t.dir)
}

// It should return an empty string if there is no current uuid in the
// repository.
func (t *TransactionContainerManagerTest) TestEmptyCurrentUUID(c *C) {
	currentUUID, err := t.tcm.currentTransactionContainerUUID()
	c.Assert(err, IsNil)
	c.Assert(currentUUID, Equals, "")
}

// It should create a new container without any other container being
// created yet if the current container is requested.
func (t *TransactionContainerManagerTest) TestNewContainerCreation(c *C) {
	transactionContainer, err := t.tcm.CurrentTransactionContainer()
	c.Assert(err, IsNil)
	c.Assert(transactionContainer.UUID, Not(Equals), "")
	c.Assert(transactionContainer.PreviousUUID, Equals, "")
	c.Assert(transactionContainer.Transactions, HasLen, 0)
}

// It should add the new transaction manager to the repository if it is initially
// created via the CurrentTransactionContainer method.
func (t *TransactionContainerManagerTest) TestNewContainerAddition(c *C) {
	transactionContainer, err := t.tcm.CurrentTransactionContainer()
	c.Assert(err, IsNil)
	exists := t.tcm.Exists(transactionContainer.UUID)
	c.Assert(exists, Equals, true)
}

// It should return the same container if the CurrentTransactionContainer is querried
// twice.
func (t *TransactionContainerManagerTest) TestNewContainerStorage(c *C) {
	transactionContainer, err := t.tcm.CurrentTransactionContainer()
	c.Assert(err, IsNil)

	otherTransactionContainer, err := t.tcm.CurrentTransactionContainer()
	c.Assert(err, IsNil)
	c.Assert(transactionContainer.UUID, Equals, otherTransactionContainer.UUID)
}

// It should always return the newest set containerStorage as the current one.
func (t *TransactionContainerManagerTest) TestCreateNew(c *C) {
	transactionContainer, err := t.tcm.CurrentTransactionContainer()
	c.Assert(err, IsNil)

	newContainer, err := t.tcm.NewContainer()
	c.Assert(err, IsNil)

	transactionContainer, err = t.tcm.CurrentTransactionContainer()

	c.Assert(newContainer.UUID, Equals, transactionContainer.UUID)
}

func (t *TransactionContainerManagerTest) SetTransactionContainer(c *C) *TransactionContainer {
	transactionContainer := &TransactionContainer{
		UUID:         "testinit",
		Transactions: []*Transaction{},
		PreviousUUID: ""}
	err := t.tcm.Set(transactionContainer)
	c.Assert(err, IsNil)
	return transactionContainer
}

// It should be able to set a new transactionContainer.
func (t *TransactionContainerManagerTest) TestSet(c *C) {
	t.SetTransactionContainer(c)
}

// It should always return the newest set entry.
func (t *TransactionContainerManagerTest) TestSetCurrent(c *C) {
	transactionContainer := t.SetTransactionContainer(c)

	currentTransactionContainer, err := t.tcm.CurrentTransactionContainer()

	c.Assert(err, IsNil)
	c.Assert(transactionContainer.UUID, Equals, currentTransactionContainer.UUID)
}

// It should return a FileNotExists if the container is not existing.
func (t *TransactionContainerManagerTest) TestGetNegative(c *C) {
	_, err := t.tcm.Get("doesnotexist")
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (t *TransactionContainerManagerTest) TestGet(c *C) {
	transactionContainer := t.SetTransactionContainer(c)
	retTransactionContainer, err := t.tcm.Get(transactionContainer.UUID)
	c.Assert(err, IsNil)
	c.Assert(retTransactionContainer.UUID, Equals, transactionContainer.UUID)
}

func (t *TransactionContainerManagerTest) TestStoreTransaction(c *C) {
	transactionContainer := &TransactionContainer{
		UUID: "testinit",
		Transactions: []*Transaction{
			&Transaction{
				ID:         1,
				NIBIDs:     []string{"a", "b", "c"},
				PreviousID: 0},
			&Transaction{
				ID:         2,
				NIBIDs:     []string{"d", "e", "f"},
				PreviousID: 1}},
		PreviousUUID: ""}
	err := t.tcm.Set(transactionContainer)
	c.Assert(err, IsNil)

	retTransactionContainer, err := t.tcm.Get(transactionContainer.UUID)
	c.Assert(err, IsNil)

	c.Assert(retTransactionContainer.UUID, Equals, transactionContainer.UUID)
	c.Assert(len(retTransactionContainer.Transactions), Equals, 2)

	for index, transaction := range retTransactionContainer.Transactions {
		checkTransaction := transactionContainer.Transactions[index]
		c.Assert(transaction.ID, Equals, checkTransaction.ID)
		c.Assert(transaction.NIBIDs, DeepEquals, checkTransaction.NIBIDs)
		c.Assert(transaction.PreviousID, Equals, checkTransaction.PreviousID)
	}
}

// It should raise an error if the ID of a Transaction in the container
// is set to 0.
func (t *TransactionContainerManagerTest) TestStoreTransactionIDZero(c *C) {
	transactionContainer := &TransactionContainer{
		UUID: "testinit",
		Transactions: []*Transaction{
			&Transaction{
				ID:         0,
				NIBIDs:     []string{"a", "b", "c"},
				PreviousID: 0,
			},
		},
		PreviousUUID: ""}

	err := t.tcm.Set(transactionContainer)
	c.Assert(err, NotNil)
}

// It should not be able to set a Transaction which does not any NIBUUIDs stored.
func (t *TransactionContainerManagerTest) TestStoreTransactionEmptyNIBID(c *C) {
	transactionContainer := &TransactionContainer{
		UUID: "testinit",
		Transactions: []*Transaction{
			&Transaction{
				ID:         1,
				NIBIDs:     []string{},
				PreviousID: 0,
			},
		},
		PreviousUUID: ""}

	err := t.tcm.Set(transactionContainer)
	c.Assert(err, NotNil)
}

// It should return False if the given entry does not exist.
func (t *TransactionContainerManagerTest) TestExistsNegative(c *C) {
	c.Assert(t.tcm.Exists("testinit"), Equals, false)
}

// It should return True if the given entry does exist.
func (t *TransactionContainerManagerTest) TestExistsPositive(c *C) {
	t.SetTransactionContainer(c)
	c.Assert(t.tcm.Exists("testinit"), Equals, true)
}
