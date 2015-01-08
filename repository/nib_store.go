package repository

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
)

var (
	// ErrSignatureVerification gets returned if a signature of a signed NIB could
	// not be verified.
	ErrSignatureVerification = errors.New("Signature verification failed")
	// ErrUnMarshalling gets returned if a NIB could not be extracted from stored
	// bytes.
	ErrUnMarshalling = errors.New("Couldn't extract item from byte stream")
)

// NIBStore handles the interaction with NIBs in a specific
// repository.
type NIBStore struct {
	storage            *UUIDContentStorage
	repository         *Repository
	transactionManager *TransactionManager
}

// newNibStore generates the NIBStore with the passed backend, repository,
// and transactionManager.
func newNIBStore(
	storage *ContentStorage,
	repository *Repository,
	transactionManager *TransactionManager,
) *NIBStore {
	uuidStorage := UUIDContentStorage{*storage}
	return &NIBStore{
		storage:            &uuidStorage,
		repository:         repository,
		transactionManager: transactionManager,
	}
}

// Get returns the NIB of the given uuid.
func (s *NIBStore) Get(UUID string) (*NIB, error) {
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

	if !signatureReader.VerifyAfterRead() {
		return nil, ErrSignatureVerification
	}

	return &nib, nil
}

// GetBytes returns the Byte representation of the
// given NIB UUID.
func (s *NIBStore) GetBytes(UUID string) ([]byte, error) {
	reader, err := s.GetReader(UUID)
	if err != nil {
		return []byte{}, err
	}
	return ioutil.ReadAll(reader)
}

// GetReader returns the Reader which stores the bytes
// of the given NIB UUID.
func (s *NIBStore) GetReader(UUID string) (io.Reader, error) {
	return s.storage.Get(UUID)
}

func (s *NIBStore) getFromTransactions(transactions []*Transaction) <-chan *NIB {
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
func (s *NIBStore) GetAll() (<-chan *NIB, error) {
	transactions, err := s.transactionManager.All()
	if err != nil {
		return nil, err
	}

	return s.getFromTransactions(transactions), nil
}

// GetFrom returns all NIBs added added after the given transaction id.
func (s *NIBStore) GetFrom(transactionID int64) (<-chan *NIB, error) {
	transactions, err := s.transactionManager.From(transactionID)
	if err != nil {
		return nil, err
	}

	return s.getFromTransactions(transactions), nil
}

// Add adds the given NIB to the store.
func (s *NIBStore) Add(nib *NIB) error {
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
func (s *NIBStore) writeBytes(UUID string, data []byte) error {
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
func (s *NIBStore) createTransaction(UUID string) *Transaction {
	return &Transaction{
		NIBUUIDs: []string{UUID},
	}
}

// AddContent adds the byte data of a NIB with the passed UUID in the
// storage backend.
func (s *NIBStore) AddContent(UUID string, reader io.Reader) error {
	transaction := s.createTransaction(UUID)

	err := s.storage.Set(UUID, reader)
	if err != nil {
		return err
	}
	return s.transactionManager.Add(transaction)
}

// Exists returns if there is a NIB with
// the given UUID in the store.
func (s *NIBStore) Exists(UUID string) bool {
	return s.storage.Exists(UUID)
}

// VerifyContent verifies the correctness of the given
// data in the reader.
func (s *NIBStore) VerifyContent(reader io.Reader) error {
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
