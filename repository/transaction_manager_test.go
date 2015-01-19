package repository

import (
	"github.com/hoffie/larasync/repository/content"

	. "gopkg.in/check.v1"
)

type TransactionManagerTest struct {
	dir string
	tm  *TransactionManager
}

var _ = Suite(&TransactionManagerTest{})

func (t *TransactionManagerTest) SetUpTest(c *C) {
	t.dir = c.MkDir()
	storage := content.NewFileStorage(t.dir)

	t.tm = newTransactionManager(storage, t.dir)
}

func (t *TransactionManagerTest) transactions(count int) []*Transaction {
	nibIDs := []string{"a", "b"}
	transactions := make([]*Transaction, count)
	for i := 1; i <= count; i++ {
		transaction := &Transaction{
			ID:     int64(i),
			NIBIDs: nibIDs}
		transactions[i-1] = transaction
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
		ID:     1,
		NIBIDs: []string{"a", "b", "c"}}
	err := t.tm.Add(transaction)
	c.Assert(err, IsNil)
}

func (t *TransactionManagerTest) TestAddTransactionFirst(c *C) {
	transaction := &Transaction{
		ID:     1,
		NIBIDs: []string{"a", "b", "c"}}
	err := t.tm.Add(transaction)
	c.Assert(err, IsNil)
	c.Assert(transaction.PreviousID, Equals, int64(0))
}

func (t *TransactionManagerTest) TestAddTransactionPreviousSet(c *C) {
	transaction := &Transaction{
		ID:     1,
		NIBIDs: []string{"a", "b"}}
	err := t.tm.Add(transaction)
	c.Assert(err, IsNil)

	transaction = &Transaction{
		ID:     2,
		NIBIDs: []string{"c", "d"}}
	err = t.tm.Add(transaction)
	c.Assert(err, IsNil)

	c.Assert(transaction.PreviousID, Equals, int64(1))
}

func (t *TransactionManagerTest) TestGetInFirstSet(c *C) {
	t.addTransactions()
	transaction, err := t.tm.Get(150)
	c.Assert(err, IsNil)
	c.Assert(transaction.ID, Equals, int64(150))
	c.Assert(transaction.PreviousID, Equals, int64(149))
}

func (t *TransactionManagerTest) TestGetInSecondSet(c *C) {
	t.addTransactions()
	transaction, err := t.tm.Get(1)
	c.Assert(err, IsNil)
	c.Assert(transaction.ID, Equals, int64(1))
	c.Assert(transaction.PreviousID, Equals, int64(0))
	_, err = t.tm.Get(0)
	c.Assert(err, NotNil)
}

func (t *TransactionManagerTest) TestFromFirstSetOnly(c *C) {
	t.addTransactions()
	transactions, err := t.tm.From(151)
	c.Assert(err, IsNil)
	c.Assert(len(transactions), Equals, 49)
}

func (t *TransactionManagerTest) TestFromSets(c *C) {
	t.addTransactions()
	transactions, err := t.tm.From(1)
	c.Assert(err, IsNil)
	c.Assert(transactions[0].ID, Equals, int64(2))
	c.Assert(len(transactions), Equals, 199)
	c.Assert(transactions[198].ID, Equals, int64(200))
}

func (t *TransactionManagerTest) TestFromNotIncludeGivenUUID(c *C) {
	t.addTransactions()
	transactions, err := t.tm.From(150)
	c.Assert(err, IsNil)
	for _, transaction := range transactions {
		c.Assert(transaction.ID, Not(Equals), int64(150))
	}
}

func (t *TransactionManagerTest) TestExistsPositive(c *C) {
	t.addTransactions()
	c.Assert(t.tm.Exists(2), Equals, true)
}

func (t *TransactionManagerTest) TestExistsNegative(c *C) {
	t.addTransactions()
	queryID := int64(transactionsInContainer*2 + 1)
	c.Assert(t.tm.Exists(queryID), Equals, false)
}
