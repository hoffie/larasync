package repository

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/hoffie/larasync/repository/odf"
)

// TransactionContainerManager is used to manage the transaction containers
// and to keep track of the most current transaction manager,
type TransactionContainerManager struct {
	storage UUIDContentStorage
	mutex   *sync.Mutex
}

// newTransactionContainerManager initializes a container manager
// the passed content storage which is used to access the stored
// data entries.
func newTransactionContainerManager(storage ContentStorage) *TransactionContainerManager {
	uuidStorage := UUIDContentStorage{storage}
	return &TransactionContainerManager{
		storage: uuidStorage,
		//@TODO: Implement global state which returns one mutex for the repository.
		mutex: &sync.Mutex{},
	}
}

// Get returns the TransactionContainer with the given UUID.
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

	transactionContainer := newTransactionContainerFromPb(protoTransactionContainer)
	return transactionContainer, nil
}

// Set sets the transactionContainer in the storage backend.
func (tcm TransactionContainerManager) Set(transactionContainer *TransactionContainer) error {
	if transactionContainer.UUID == "" {
		return errors.New("UUID must not be empty")
	}
	mutex := tcm.mutex

	mutex.Lock()
	err := func() error {
		protoTransactionContainer, err := transactionContainer.toPb()
		if err != nil {
			return err
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
	}()
	mutex.Unlock()
	return err
}

// Exists returns if a TransactionContainer with the given UUID exists in the system.
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
		}
		return "", err
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
	}
	// Have to create a new TransactionContainer due to no current existing yet.
	return tcm.NewContainer()
}

// NewContainer returns a newly container with a new UUID which has been added to the
// storage backend.
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