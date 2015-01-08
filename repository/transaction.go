package repository

import (
	"errors"
	"fmt"

	"github.com/hoffie/larasync/repository/odf"
)

// Transaction represents a server side transaction for specific NIBs
// which is used to synchronize the different clients.
type Transaction struct {
	ID         int64
	NIBUUIDs   []string
	PreviousID int64
}

// newTransactionFromPb returns a new Transaction from the
// protobuf Transaction.
func newTransactionFromPb(pbTransaction *odf.Transaction) *Transaction {
	return &Transaction{
		ID:         pbTransaction.GetID(),
		PreviousID: pbTransaction.GetPreviousID(),
		NIBUUIDs:   pbTransaction.GetNIBUUIDs(),
	}
}

// toPb converts this Transaction to a protobuf Transaction.
// This is used by the encoder.
func (t *Transaction) toPb() (*odf.Transaction, error) {
	if t.ID == 0 {
		return nil, errors.New("Transaction ID must not be empty")
	}
	if len(t.NIBUUIDs) == 0 {
		return nil, fmt.Errorf(
			"The transition with ID %d has no NIB UUIDs",
			t.ID,
		)
	}
	protoTransaction := &odf.Transaction{
		ID:         &t.ID,
		PreviousID: nil,
		NIBUUIDs:   t.NIBUUIDs}
	if t.PreviousID != 0 {
		protoTransaction.PreviousID = &t.PreviousID
	}
	return protoTransaction, nil
}
