package repository

import (
	"fmt"

	. "gopkg.in/check.v1"
)

type TransactionManagerTest struct {
	dir string
	tm  *TransactionManager
}

var _ = Suite(&TransactionManagerTest{})

func (t *TransactionManagerTest) SetUpTest(c *C) {
	t.dir = c.MkDir()
	storage := &FileContentStorage{StoragePath: t.dir}

	t.tm = newTransactionManager(storage)
}

func (t *TransactionManagerTest) transactions(count int) []*Transaction {
	nibUUIDs := []string{"a", "b"}
	transactions := make([]*Transaction, count)
	for i := 0; i < count; i++ {
		transaction := &Transaction{
			UUID:     fmt.Sprintf("uuid%d", i),
			NIBUUIDs: nibUUIDs}
		transactions[i] = transaction
	}
	return transactions
}

func (t *TransactionManagerTest) addTransactions() {
	transactions := t.transactions(transactionsInContainer * 2)
	for _, transaction := range transactions {
		t.tm.Add(transaction)
	}
}

func (t *TransactionManagerTest) TestAddTransaction(c *C) {
	transaction := &Transaction{
		UUID:     "uuid1",
		NIBUUIDs: []string{"a", "b", "c"}}
	err := t.tm.Add(transaction)
	c.Assert(err, IsNil)
}

func (t *TransactionManagerTest) TestAddTransactionFirst(c *C) {
	transaction := &Transaction{
		UUID:     "uuid1",
		NIBUUIDs: []string{"a", "b", "c"}}
	err := t.tm.Add(transaction)
	c.Assert(err, IsNil)
	c.Assert(transaction.PreviousUUID, Equals, "")
}

func (t *TransactionManagerTest) TestAddTransactionPreviousSet(c *C) {
	transaction := &Transaction{
		UUID:     "uuid1",
		NIBUUIDs: []string{"a", "b"}}
	err := t.tm.Add(transaction)
	c.Assert(err, IsNil)

	transaction = &Transaction{
		UUID:     "uuid2",
		NIBUUIDs: []string{"c", "d"}}
	err = t.tm.Add(transaction)
	c.Assert(err, IsNil)

	c.Assert(transaction.PreviousUUID, Equals, "uuid1")
}

func (t *TransactionManagerTest) TestGetInFirstSet(c *C) {
	t.addTransactions()
	transaction, err := t.tm.Get("uuid150")
	c.Assert(err, IsNil)
	c.Assert(transaction.UUID, Equals, "uuid150")
	c.Assert(transaction.PreviousUUID, Equals, "uuid149")
}

func (t *TransactionManagerTest) TestGetInSecondSet(c *C) {
	t.addTransactions()
	transaction, err := t.tm.Get("uuid0")
	c.Assert(err, IsNil)
	c.Assert(transaction.UUID, Equals, "uuid0")
	c.Assert(transaction.PreviousUUID, Equals, "")
}

func (t *TransactionManagerTest) TestFromFirstSetOnly(c *C) {
	t.addTransactions()
	transactions, err := t.tm.From("uuid150")
	c.Assert(err, IsNil)
	c.Assert(len(transactions), Equals, 49)
}

func (t *TransactionManagerTest) TestFromSets(c *C) {
	t.addTransactions()
	transactions, err := t.tm.From("uuid1")
	c.Assert(err, IsNil)
	c.Assert(transactions[0].UUID, Equals, "uuid2")
	c.Assert(len(transactions), Equals, 198)
	c.Assert(transactions[197].UUID, Equals, "uuid199")
}

func (t *TransactionManagerTest) TestFromNotIncludeGivenUUID(c *C) {
	t.addTransactions()
	transactions, err := t.tm.From("uuid150")
	c.Assert(err, IsNil)
	for _, transaction := range transactions {
		c.Assert(transaction, Not(Equals), "uuid150")
	}
}

func (t *TransactionManagerTest) TestExistsPositive(c *C) {
	t.addTransactions()
	c.Assert(t.tm.Exists("uuid2"), Equals, true)
}

func (t *TransactionManagerTest) TestExistsNegative(c *C) {
	t.addTransactions()
	c.Assert(t.tm.Exists(fmt.Sprintf("uuid%d", transactionsInContainer*2)), Equals, false)
}
