package repository

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
)

// ClientNIBStore implements the NIBStore interface from the
// client perspective.
type ClientNIBStore struct {
	storage            *UUIDContentStorage
	repository         *Repository
	transactionManager *TransactionManager
}

// newClientNIBStore generates the clientNibStore with the given data
// and returns the new entry.
func newClientNIBStore(
	storage *ContentStorage,
	repository *Repository,
	transactionManager *TransactionManager,
) *ClientNIBStore {
	nibStorage := &UUIDContentStorage{*storage}
	return &ClientNIBStore{
		storage:            nibStorage,
		repository:         repository,
		transactionManager: transactionManager,
	}
}

// Get returns the NIB of the given id.
func (s ClientNIBStore) Get(id string) (*NIB, error) {
	pubKey, err := s.repository.GetSigningPubkey()
	if err != nil {
		return nil, err
	}

	data, err := s.GetBytes(id)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewReader(data)
	signatureReader, err := NewVerifyingReader(
		pubKey,
		buffer,
	)
	if err != nil {
		return nil, err
	}

	nib := NIB{}
	_, err = nib.ReadFrom(signatureReader)
	if err != nil {
		return nil, err
	}

	if !signatureReader.VerifyAfterRead() {
		return nil, ErrSignatureVerification
	}

	return &nib, nil
}

// GetBytes returns the Byte representation of the
// given NIB ID.
func (s ClientNIBStore) GetBytes(id string) ([]byte, error) {
	reader, err := s.GetReader(id)
	if err != nil {
		return []byte{}, err
	}
	return ioutil.ReadAll(reader)
}

// GetReader returns the Reader which stores the bytes
// of the given NIB ID.
func (s ClientNIBStore) GetReader(id string) (io.Reader, error) {
	return s.storage.Get(id)
}

func (s ClientNIBStore) getFromTransactions(transactions []*Transaction) <-chan *NIB {
	nibChannel := make(chan *NIB, 100)

	go func() {
		for _, transaction := range transactions {
			for _, nibID := range transaction.NIBIDs {
				nib, err := s.Get(nibID)
				if err != nil {
					break
				}
				nibChannel <- nib
			}
		}
		close(nibChannel)
	}()

	return nibChannel
}

// GetAll returns all NIBs which have been commited to the store.
func (s ClientNIBStore) GetAll() (<-chan *NIB, error) {
	transactions, err := s.transactionManager.All()
	if err != nil {
		return nil, err
	}

	return s.getFromTransactions(transactions), nil
}

// GetFrom returns all NIBs added added after the given transaction id.
func (s ClientNIBStore) GetFrom(transactionID int64) (<-chan *NIB, error) {
	transactions, err := s.transactionManager.From(transactionID)
	if err != nil {
		return nil, err
	}

	return s.getFromTransactions(transactions), nil
}

// Add adds the given NIB to the store.
func (s ClientNIBStore) Add(nib *NIB) error {
	if nib.ID == "" {
		return errors.New("empty NIB id")
	}

	buf := &bytes.Buffer{}
	_, err := nib.WriteTo(buf)
	if err != nil {
		return err
	}

	return s.writeBytes(nib.ID, buf.Bytes())
}

// writeBytes signs and adds the bytes for the given NIB ID.
func (s ClientNIBStore) writeBytes(id string, data []byte) error {
	key, err := s.repository.GetSigningPrivkey()

	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}

	sw := NewSigningWriter(key, buf)
	_, err = sw.Write(data)
	if err != nil {
		return err
	}
	err = sw.Finalize()
	if err != nil {
		return err
	}

	return s.AddContent(id, buf)
}

// Creates a transaction with the given id as NIB and Transaction id.
func (s ClientNIBStore) createTransaction(id string) *Transaction {
	return &Transaction{
		NIBIDs: []string{id},
	}
}

func (s ClientNIBStore) AddContent(id string, reader io.Reader) error {
	transaction := s.createTransaction(id)

	err := s.storage.Set(id, reader)
	if err != nil {
		return err
	}
	return s.transactionManager.Add(transaction)
}

// Exists returns if there is a NIB with
// the given ID in the store.
func (s ClientNIBStore) Exists(id string) bool {
	return s.storage.Exists(id)
}

// VerifyContent verifies the correctness of the given
// data in the reader.
func (s ClientNIBStore) VerifyContent(reader io.Reader) error {
	pubKey, err := s.repository.GetSigningPubkey()
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	signatureReader, err := NewVerifyingReader(
		pubKey,
		bytes.NewReader(data),
	)
	if err != nil {
		return err
	}

	nib := &NIB{}
	_, err = nib.ReadFrom(signatureReader)
	if err != nil {
		return ErrUnMarshalling
	}

	if !signatureReader.VerifyAfterRead() {
		return ErrSignatureVerification
	}

	return nil
}
