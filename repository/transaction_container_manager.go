package repository

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"

	"github.com/hoffie/larasync/repository/odf"
)

// TransactionContainerManager is used to manage the transaction containers
// and to keep track of the most current transaction manager,
type TransactionContainerManager struct {
	storage UUIDContentStorage
}

// newTransactionContainerManager initializes a container manager
// the passed content storage which is used to access the stored
// data entries.
func newTransactionContainerManager(storage ContentStorage) *TransactionContainerManager {
	uuidStorage := UUIDContentStorage{storage}
	return &TransactionContainerManager{storage: uuidStorage}
}

func (tcm TransactionContainerManager) Get(transactionContainerUUID string) (*TransactionContainer, error) {
	reader, err := tcm.storage.Get(transactionContainerUUID)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	protoTransactionContainer := &odf.TransactionContainer{}
	err = proto.Unmarshal(
		data,
		protoTransactionContainer)

	if err != nil {
		return nil, err
	}

	transactions := make([]*Transaction, len(protoTransactionContainer.Transactions))
	transactionContainer := &TransactionContainer{
		UUID:         *protoTransactionContainer.UUID,
		PreviousUUID: "",
		Transactions: transactions}
	if protoTransactionContainer.PreviousUUID != nil {
		transactionContainer.PreviousUUID = *protoTransactionContainer.PreviousUUID
	}

	for index, protoTransaction := range protoTransactionContainer.Transactions {
		uuid := *protoTransaction.UUID
		transaction := &Transaction{
			UUID:         uuid,
			PreviousUUID: "",
			NIBUUIDs:     []string{}}
		if protoTransaction.PreviousUUID != nil {
			transaction.PreviousUUID = *protoTransaction.PreviousUUID
		}
		transaction.NIBUUIDs = protoTransaction.NIBUUIDs

		transactions[index] = transaction
	}

	return transactionContainer, nil
}

func (tcm TransactionContainerManager) Set(transactionContainer *TransactionContainer) error {
	if transactionContainer.UUID == "" {
		return errors.New("UUID must not be empty")
	}

	previousUUID := ""
	protoTransactions := make([]*odf.Transaction, len(transactionContainer.Transactions))
	protoTransactionContainer := &odf.TransactionContainer{
		UUID:         &transactionContainer.UUID,
		PreviousUUID: &previousUUID,
		Transactions: protoTransactions}

	if transactionContainer.PreviousUUID != "" {
		protoTransactionContainer.PreviousUUID = &transactionContainer.PreviousUUID
	}

	for index, transaction := range transactionContainer.Transactions {
		if transaction.UUID == "" {
			return errors.New("Transaction UUID must not be empty")
		}
		if len(transaction.NIBUUIDs) == 0 {
			return fmt.Errorf("The transition with UUID %s has no NIB UUIDs",
				transaction.UUID)
		}

		protoTransaction := &odf.Transaction{
			UUID:         &transaction.UUID,
			PreviousUUID: nil,
			NIBUUIDs:     transaction.NIBUUIDs}

		if transaction.PreviousUUID != "" {
			protoTransaction.PreviousUUID = &transaction.PreviousUUID
		}

		protoTransactions[index] = protoTransaction
	}

	data, err := proto.Marshal(protoTransactionContainer)
	if err != nil {
		return err
	}

	err = tcm.storage.Set(
		transactionContainer.UUID,
		bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	return tcm.storage.Set(
		"current",
		bytes.NewBufferString(transactionContainer.UUID))
}

func (tcm TransactionContainerManager) Exists(transactionContainerUUID string) bool {
	return tcm.storage.Exists(transactionContainerUUID)
}

// currentTransactionContainerUUID reads the stored currently
// configured UUID for the transaction container.
func (tcm TransactionContainerManager) currentTransactionContainerUUID() (string, error) {
	reader, err := tcm.storage.Get("current")
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		} else {
			return "", err
		}
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// CurrentTransactionContainer returns the TransactionContainer which is the most recent
// for the given repository.
func (tcm TransactionContainerManager) CurrentTransactionContainer() (*TransactionContainer, error) {
	currentUUID, err := tcm.currentTransactionContainerUUID()
	if err != nil {
		return nil, err
	}

	if currentUUID != "" {
		return tcm.Get(currentUUID)
	} else {
		// Have to create a new TransactionContainer due to no current existing yet.
		return tcm.NewContainer()
	}
}

func (tcm TransactionContainerManager) NewContainer() (*TransactionContainer, error) {
	data, err := tcm.storage.findFreeUUID()
	if err != nil {
		return nil, err
	}

	uuid := formatUUID(data)
	previousUUID, err := tcm.currentTransactionContainerUUID()
	if err != nil {
		return nil, err
	}

	transactionContainer := &TransactionContainer{
		UUID:         uuid,
		Transactions: []*Transaction{},
		PreviousUUID: previousUUID}

	err = tcm.Set(transactionContainer)
	if err != nil {
		return nil, err
	}

	return transactionContainer, nil
}
