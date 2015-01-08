package repository

import (
	"errors"
	"sync"
)

var (
	// ErrTransactionNotExists is thrown if a transaction could not be found.
	ErrTransactionNotExists = errors.New("Transaction does not exist in repository.")
)

const transactionsInContainer int = 100

func reverseTransactionSlice(slice []*Transaction) []*Transaction {
	for i := 0; i < len(slice)/2; i++ {
		slice[i], slice[len(slice)-1-i] = slice[len(slice)-1-i], slice[i]
	}
	return slice
}

// TransactionManager is used to query and add data written
// in the server transaction log.
type TransactionManager struct {
	manager *TransactionContainerManager
	mutex   *sync.Mutex
}

// newTransactionManager initializes a new transaction manager
// with the given storage as a backend.
func newTransactionManager(storage ContentStorage) *TransactionManager {
	manager := newTransactionContainerManager(storage)
	return &TransactionManager{
		manager: manager,
		mutex:   &sync.Mutex{},
	}
}

// CurrentTransactionID returns the most recent ID stored in the
// backend.
func (tm *TransactionManager) CurrentTransactionID() (int64, error) {
	newestTransaction, err := tm.CurrentTransaction()
	if err != nil {
		return 0, err
	}
	return newestTransaction.ID, nil
}

// CurrentTransaction returns the most recent Transaction which is stored
// in the TransactionLog.
func (tm *TransactionManager) CurrentTransaction() (*Transaction, error) {
	currentTransactionContainer, err := tm.manager.CurrentTransactionContainer()
	if err != nil {
		return nil, err
	}

	transactionsLength := len(currentTransactionContainer.Transactions)
	if transactionsLength == 0 {
		// Empty transactions. No current UUID.
		return nil, ErrTransactionNotExists
	}

	transactions := currentTransactionContainer.Transactions
	newestTransaction := transactions[transactionsLength-1]
	return newestTransaction, nil
}

// Add adds the given transaction to the storage.
func (tm *TransactionManager) Add(transaction *Transaction) error {
	mutex := tm.mutex

	mutex.Lock()
	err := func() error {
		manager := tm.manager
		transactionContainer, err := manager.CurrentTransactionContainer()
		if err != nil {
			return err
		}

		var previousID int64
		if len(transactionContainer.Transactions) > 0 {
			latestIndex := len(transactionContainer.Transactions) - 1
			previousID = transactionContainer.Transactions[latestIndex].ID
		}

		if len(transactionContainer.Transactions) >= transactionsInContainer {
			transactionContainer, err = manager.NewContainer()
			if err != nil {
				return err
			}
		}

		transaction.PreviousID = previousID
		if transaction.ID == 0 {
			transaction.ID = previousID + 1
		}

		transactionContainer.Transactions = append(
			transactionContainer.Transactions,
			transaction,
		)
		return manager.Set(transactionContainer)
	}()
	mutex.Unlock()
	return err
}

// Get returns the transaction with the given UUID.
func (tm *TransactionManager) Get(transactionID int64) (*Transaction, error) {
	manager := tm.manager
	currentTransactionContainer, err := manager.CurrentTransactionContainer()
	if err != nil {
		return nil, err
	}

	transactionContainer := currentTransactionContainer
	for transactionContainer != nil {
		for _, transaction := range transactionContainer.Transactions {
			if transaction.ID == transactionID {
				return transaction, nil
			}
		}

		if transactionContainer.PreviousUUID == "" {
			transactionContainer = nil
		} else {
			transactionContainer, err = manager.Get(
				transactionContainer.PreviousUUID)
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, ErrTransactionNotExists

}

// From returns all transactions from the given transactionUUID. It does not include
// the transaction of the given transactionUUID.
func (tm *TransactionManager) From(transactionID int64) ([]*Transaction, error) {
	manager := tm.manager
	currentTransactionContainer, err := manager.CurrentTransactionContainer()
	if err != nil {
		return nil, err
	}

	transactions := [][]*Transaction{}
	transactionContainer := currentTransactionContainer
	found := false
	for transactionContainer != nil {
		foundTransactions := []*Transaction{}
		for _, transaction := range transactionContainer.Transactions {
			if found {
				foundTransactions = append(foundTransactions, transaction)
			}

			if transaction.ID == transactionID {
				found = true
			}
		}
		if found {
			transactions = append(transactions, foundTransactions)
			break
		}

		transactions = append(transactions, transactionContainer.Transactions)

		if transactionContainer.PreviousUUID == "" {
			transactionContainer = nil
		} else {
			transactionContainer, err = manager.Get(
				transactionContainer.PreviousUUID)
			if err != nil {
				return nil, err
			}
		}
	}

	returnTransactions := []*Transaction{}
	for i := len(transactions) - 1; i >= 0; i-- {
		returnTransactions = append(returnTransactions, transactions[i]...)
	}

	return returnTransactions, nil
}

// All returns all transactions in the system.
func (tm *TransactionManager) All() ([]*Transaction, error) {
	return tm.From(0)
}

// Exists checks if the given Transaction UUID exists in this repository.
func (tm *TransactionManager) Exists(transactionID int64) bool {
	_, err := tm.Get(transactionID)
	if err != nil {
		return false
	}
	return true
}
