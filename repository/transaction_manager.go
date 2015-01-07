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

// CurrentTransactionUUID returns the most recent UUID stored in the
// backend.
func (tm *TransactionManager) CurrentTransactionUUID() (string, error) {
	newestTransaction, err := tm.CurrentTransaction()
	if err != nil {
		return "", err
	}
	return newestTransaction.UUID, nil
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
	newestTransaction := transactions[transactionsLength]
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

		var previousUUID string
		if len(transactionContainer.Transactions) > 0 {
			latestIndex := len(transactionContainer.Transactions) - 1
			previousUUID = transactionContainer.Transactions[latestIndex].UUID
		}

		if len(transactionContainer.Transactions) >= transactionsInContainer {
			transactionContainer, err = manager.NewContainer()
			if err != nil {
				return err
			}
		}

		transaction.PreviousUUID = previousUUID
		transactionContainer.Transactions = append(
			transactionContainer.Transactions,
			transaction)
		return manager.Set(transactionContainer)
	}()
	mutex.Unlock()
	return err
}

// Get returns the transaction with the given UUID.
func (tm *TransactionManager) Get(transactionUUID string) (*Transaction, error) {
	manager := tm.manager
	currentTransactionContainer, err := manager.CurrentTransactionContainer()
	if err != nil {
		return nil, err
	}

	transactionContainer := currentTransactionContainer
	for transactionContainer != nil {
		for _, transaction := range transactionContainer.Transactions {
			if transaction.UUID == transactionUUID {
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
func (tm *TransactionManager) From(transactionUUID string) ([]*Transaction, error) {
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

			if transaction.UUID == transactionUUID {
				found = true
			}
		}
		if found {
			transactions = append(transactions, foundTransactions)
			returnTransactions := []*Transaction{}
			for i := len(transactions) - 1; i >= 0; i-- {
				returnTransactions = append(returnTransactions, transactions[i]...)
			}

			return returnTransactions, nil
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

	return nil, ErrTransactionNotExists
}

// Exists checks if the given Transaction UUID exists in this repository.
func (tm *TransactionManager) Exists(transactionUUID string) bool {
	_, err := tm.Get(transactionUUID)
	if err != nil {
		return false
	}
	return true
}
