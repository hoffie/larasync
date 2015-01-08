package repository

import (
	"bytes"
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
	uuidStorage := UUIDContentStorage{*storage}
	return &ClientNIBStore{
		storage:            &uuidStorage,
		repository:         repository,
		transactionManager: transactionManager,
	}
}

// Get returns the NIB of the given uuid.
func (s ClientNIBStore) Get(UUID string) (*NIB, error) {
	pubKey, err := s.repository.GetSigningPubkey()
	if err != nil {
		return nil, err
	}

	data, err := s.GetBytes(UUID)
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

	if !signatureReader.Verify() {
		return nil, SignatureVerificationError
	}

	return &nib, nil
}

// GetBytes returns the Byte representation of the
// given NIB UUID.
func (s ClientNIBStore) GetBytes(UUID string) ([]byte, error) {
	reader, err := s.GetReader(UUID)
	if err != nil {
		return []byte{}, err
	}
	return ioutil.ReadAll(reader)
}

// GetReader returns the Reader which stores the bytes
// of the given NIB UUID.
func (s ClientNIBStore) GetReader(UUID string) (io.Reader, error) {
	return s.storage.Get(UUID)
}

func (s ClientNIBStore) getFromTransactions(transactions []*Transaction) <-chan *NIB {
	nibChannel := make(chan *NIB, 100)

	go func() {
		for _, transaction := range transactions {
			for _, nibUUID := range transaction.NIBUUIDs {
				nib, err := s.Get(nibUUID)
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
	// Empty UUID. Generating new one.
	if nib.UUID == "" {
		uuid, err := s.storage.findFreeUUID()
		if err != nil {
			return err
		}
		nib.UUID = formatUUID(uuid)
	}

	buf := &bytes.Buffer{}
	_, err := nib.WriteTo(buf)
	if err != nil {
		return err
	}

	return s.writeBytes(nib.UUID, buf.Bytes())
}

// writeBytes signs and adds the bytes for the given NIB UUID.
func (s ClientNIBStore) writeBytes(UUID string, data []byte) error {
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

	return s.AddContent(UUID, buf)
}

// Creates a transaction with the given UUID as NIB and Transaction id.
func (s ClientNIBStore) createTransaction(UUID string) *Transaction {
	return &Transaction{
		NIBUUIDs: []string{UUID},
	}
}

func (s ClientNIBStore) AddContent(UUID string, reader io.Reader) error {
	transaction := s.createTransaction(UUID)

	err := s.storage.Set(UUID, reader)
	if err != nil {
		return err
	}
	return s.transactionManager.Add(transaction)
}

// Exists returns if there is a NIB with
// the given UUID in the store.
func (s ClientNIBStore) Exists(UUID string) bool {
	return s.storage.Exists(UUID)
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
		return UnMarshallingError
	}

	if !signatureReader.Verify() {
		return SignatureVerificationError
	}

	return nil
}
