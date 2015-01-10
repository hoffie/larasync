package repository

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hoffie/larasync/repository/odf"
)

// Transaction represents a server side transaction for specific NIBs
// which is used to synchronize the different clients.
type Transaction struct {
	ID         int64
	NIBIDs     []string
	PreviousID int64
}

// newTransactionFromPb returns a new Transaction from the
// protobuf Transaction.
func newTransactionFromPb(pbTransaction *odf.Transaction) *Transaction {
	return &Transaction{
		ID:         pbTransaction.GetID(),
		PreviousID: pbTransaction.GetPreviousID(),
		NIBIDs:     pbTransaction.GetNIBIDs(),
	}
}

// toPb converts this Transaction to a protobuf Transaction.
// This is used by the encoder.
func (t *Transaction) toPb() (*odf.Transaction, error) {
	if t.ID == 0 {
		return nil, errors.New("Transaction ID must not be empty")
	}
	if len(t.NIBIDs) == 0 {
		return nil, fmt.Errorf(
			"The transaction with ID %d has no NIB IDs",
			t.ID,
		)
	}
	protoTransaction := &odf.Transaction{
		ID:         &t.ID,
		PreviousID: nil,
		NIBIDs:     t.NIBIDs}
	if t.PreviousID != 0 {
		protoTransaction.PreviousID = &t.PreviousID
	}
	return protoTransaction, nil
}

// IDString returns the ID of this Transaction as a string.
func (t *Transaction) IDString() string {
	return strconv.FormatInt(t.ID, 10)
}

// nibUUIDsFromTransactions returns all uuids from a list of transactions.
func nibUUIDsFromTransactions(transactions []*Transaction) <-chan string {
	nibUUIDChannel := make(chan string, 100)
	go func() {
		for _, transaction := range transactions {
			for _, nibID := range transaction.NIBIDs {
				nibUUIDChannel <- nibID
			}
		}
		close(nibUUIDChannel)
	}()
	return nibUUIDChannel
}
