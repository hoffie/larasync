package repository

import (
	"errors"

	"github.com/hoffie/larasync/repository/odf"
)

// TransactionContainer is used to encapsulate several transactions
// in one data format.
type TransactionContainer struct {
	UUID         string
	Transactions []*Transaction
	PreviousUUID string
}

// newRevisionFromPb returns a new TransactionContainer from the
// protobuf TransactionContainer.
func newTransactionContainerFromPb(pbTransactionContainer *odf.TransactionContainer) *TransactionContainer {
	transactions := make([]*Transaction, len(pbTransactionContainer.GetTransactions()))
	for index, protoTransaction := range pbTransactionContainer.GetTransactions() {
		transactions[index] = newTransactionFromPb(protoTransaction)
	}

	transactionContainer := &TransactionContainer{
		UUID:         pbTransactionContainer.GetUUID(),
		PreviousUUID: pbTransactionContainer.GetPreviousUUID(),
		Transactions: transactions,
	}

	return transactionContainer
}

// toPb converts this TransactionContainer to a protobuf TransactionContainer.
// This is used by the encoder.
func (tc *TransactionContainer) toPb() (*odf.TransactionContainer, error) {
	previousUUID := &tc.PreviousUUID
	if tc.PreviousUUID == "" {
		previousUUID = nil
	}

	if tc.UUID == "" {
		return nil, errors.New("UUID may not be empty")
	}

	protoTransactions := make([]*odf.Transaction, len(tc.Transactions))
	for index, transaction := range tc.Transactions {
		protoTransaction, err := transaction.toPb()
		if err != nil {
			return nil, err
		}
		protoTransactions[index] = protoTransaction
	}

	transactionContainer := &odf.TransactionContainer{
		UUID:         &tc.UUID,
		PreviousUUID: previousUUID,
		Transactions: protoTransactions,
	}

	return transactionContainer, nil
}
