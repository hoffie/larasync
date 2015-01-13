package repository

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"

	"github.com/hoffie/larasync/helpers/crypto"
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
	storage            ContentStorage
	keys               *KeyStore
	transactionManager *TransactionManager
}

// newNibStore generates the NIBStore with the passed backend, repository,
// and transactionManager.
func newNIBStore(
	storage ContentStorage,
	keys *KeyStore,
	transactionManager *TransactionManager,
) *NIBStore {
	return &NIBStore{
		storage:            storage,
		keys:               keys,
		transactionManager: transactionManager,
	}
}

// Get returns the NIB of the given id.
func (s *NIBStore) Get(id string) (*NIB, error) {
	pubKey, err := s.keys.SigningPublicKey()
	if err != nil {
		return nil, err
	}

	data, err := s.GetBytes(id)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewReader(data)
	signatureReader, err := crypto.NewVerifyingReader(
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
func (s *NIBStore) GetBytes(id string) ([]byte, error) {
	reader, err := s.getReader(id)
	if err != nil {
		return []byte{}, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}

// GetReader returns the Reader which stores the bytes
// of the given NIB ID.
func (s *NIBStore) getReader(id string) (io.ReadCloser, error) {
	return s.storage.Get(id)
}

func (s *NIBStore) getFromTransactions(transactions []*Transaction) <-chan *NIB {
	nibChannel := make(chan *NIB, 100)

	go func() {
		errorOccured := false
		for nibID := range nibUUIDsFromTransactions(transactions) {
			nib, err := s.Get(nibID)
			if err != nil {
				errorOccured = true
			}
			if !errorOccured {
				nibChannel <- nib
			}
		}
		close(nibChannel)
	}()

	return nibChannel
}

func (s *NIBStore) getByteRepresentationsFromTransactions(transactions []*Transaction) <-chan []byte {
	nibChannel := make(chan []byte, 100)
	go func() {
		for nibID := range nibUUIDsFromTransactions(transactions) {
			data, err := s.GetBytes(nibID)
			if err != nil {
				break
			}
			nibChannel <- data
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

// GetAllBytes returns all signed NIB byte representations of the NIBs
// in the repository.
func (s *NIBStore) GetAllBytes() (<-chan []byte, error) {
	transactions, err := s.transactionManager.All()
	if err != nil {
		return nil, err
	}

	return s.getByteRepresentationsFromTransactions(transactions), nil
}

// GetFrom returns all NIBs added added after the given transaction id.
func (s *NIBStore) GetFrom(transactionID int64) (<-chan *NIB, error) {
	transactions, err := s.transactionManager.From(transactionID)
	if err != nil {
		return nil, err
	}

	return s.getFromTransactions(transactions), nil
}

// GetBytesFrom returns all signed byte representations for all NIBs
// changed since the given transactionID were added.
func (s *NIBStore) GetBytesFrom(transactionID int64) (<-chan []byte, error) {
	transactions, err := s.transactionManager.From(transactionID)
	if err != nil {
		return nil, err
	}

	return s.getByteRepresentationsFromTransactions(transactions), nil
}

// Add adds the given NIB to the store.
func (s *NIBStore) Add(nib *NIB) error {
	if nib.ID == "" {
		return errors.New("empty nib ID")
	}

	buf := &bytes.Buffer{}
	_, err := nib.WriteTo(buf)
	if err != nil {
		return err
	}

	return s.writeBytes(nib.ID, buf.Bytes())
}

// writeBytes signs and adds the bytes for the given NIB ID.
func (s *NIBStore) writeBytes(id string, data []byte) error {
	key, err := s.keys.SigningPrivateKey()

	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}

	sw := crypto.NewSigningWriter(key, buf)
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

// Creates a transaction with the given ID as NIB and Transaction id.
func (s *NIBStore) createTransaction(id string) *Transaction {
	return &Transaction{
		NIBIDs: []string{id},
	}
}

// AddContent adds the byte data of a NIB with the passed ID in the
// storage backend.
func (s *NIBStore) AddContent(id string, reader io.Reader) error {
	transaction := s.createTransaction(id)

	err := s.storage.Set(id, reader)
	if err != nil {
		return err
	}
	return s.transactionManager.Add(transaction)
}

// Exists returns if there is a NIB with
// the given ID in the store.
func (s *NIBStore) Exists(id string) bool {
	return s.storage.Exists(id)
}

// VerifyAndParseBytes verifies the correctness of the given
// data in the reader and returns the parsed nib.
func (s *NIBStore) VerifyAndParseBytes(data []byte) (*NIB, error) {
	pubKey, err := s.keys.SigningPublicKey()
	if err != nil {
		return nil, err
	}

	signatureReader, err := crypto.NewVerifyingReader(
		pubKey,
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, err
	}

	// reading into a temporary buffer first requires memory,
	// but this way we avoid parsing NIBs before actually verifying
	// their signature; do not change this without further
	// consideration.
	buf, err := ioutil.ReadAll(signatureReader)
	if err != nil {
		return nil, err
	}

	if !signatureReader.VerifyAfterRead() {
		return nil, ErrSignatureVerification
	}

	nib := &NIB{}
	_, err = nib.ReadFrom(bytes.NewReader(buf))
	if err != nil {
		return nil, ErrUnMarshalling
	}

	return nib, nil
}
