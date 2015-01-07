package repository

import (
	"errors"
	"fmt"

	"github.com/hoffie/larasync/repository/odf"
)

// Transaction represents a server side transaction for specific NIBs
// which is used to synchronize the different clients.
type Transaction struct {
	UUID         string
	NIBUUIDs     []string
	PreviousUUID string
}

// newTransactionFromPb returns a new Transaction from the
// protobuf Transaction.
func newTransactionFromPb(pbTransaction *odf.Transaction) *Transaction {
	return &Transaction{
		UUID:         pbTransaction.GetUUID(),
		PreviousUUID: pbTransaction.GetPreviousUUID(),
		NIBUUIDs:     pbTransaction.GetNIBUUIDs(),
	}
}

// toPb converts this Transaction to a protobuf Transaction.
// This is used by the encoder.
func (t *Transaction) toPb() (*odf.Transaction, error) {
	if t.UUID == "" {
		return nil, errors.New("Transaction UUID must not be empty")
	}
	if len(t.NIBUUIDs) == 0 {
		return nil, fmt.Errorf(
			"The transition with UUID %s has no NIB UUIDs",
			t.UUID,
		)
	}
	protoTransaction := &odf.Transaction{
		UUID:         &t.UUID,
		PreviousUUID: nil,
		NIBUUIDs:     t.NIBUUIDs}
	if t.PreviousUUID != "" {
		protoTransaction.PreviousUUID = &t.PreviousUUID
	}
	return protoTransaction, nil
}
