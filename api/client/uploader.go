package client

import (
	"fmt"

	"github.com/hoffie/larasync/repository"
	"github.com/hoffie/larasync/repository/nib"
)

// Uploader returns the uploader for the given client in the passed
// repository.
func (c *Client) Uploader(r *repository.ClientRepository) *Uploader {
	return &Uploader{
		client: c,
		r:      r,
	}
}

// Uploader handles uploads from server to client
type Uploader struct {
	client *Client
	r      *repository.ClientRepository
}

// PushAll ensures that the remote state is synced with the local state.
func (ul *Uploader) PushAll() error {
	r := ul.r
	transaction, err := r.CurrentTransaction()
	if err != nil {
		return err
	}
	err := ul.PushAll()
	if err != nil {
		return err
	}
	return ul.saveLastUploadedTransaction(transaction)
}

// PushDelta pushes all nibs from the stored local transaction id.
func (ul *Uploader) PushDelta() error {
	r := ul.r
	s, err := r.StateConfig()
	if err != nil {
		return err
	}

	defaultServer := s.DefaultServer
	return ul.pushFromTransactionID(defaultServer.LocalTransactionID)
}

// PushFromTransactionID pushes all NIBs which have been entered after
// the given local transaction ID.
func (ul *Uploader) pushFromTransactionID(transactionID int64) error {
	r := ul.r
	transactions, err := r.TransactionsFrom(transactionID)
	if err != nil {
		return fmt.Errorf("unable to get transactions (%s)", err)
	}

	var lastTransaction *repository.Transaction
	for _, transaction := range transactions {
		err = ul.uploadTransaction(transaction)
		if err != nil {
			return err
		}
		lastTransaction = transaction
	}

	if lastTransaction != nil {
		err = ul.saveLastUploadedTransaction(lastTransaction)
		if err != nil {
			return err
		}
	}

	return nil
}

// saveLastUploadedTransaction takes the given transaction and configures it to the
// state config to store it as the last transaction.
func (ul *Uploader) saveLastUploadedTransaction(transaction *repository.Transaction) error {
	s, err := r.StateConfig()
	if err != nil {
		return nil
	}
	s.DefaultServer.LocalTransactionID = lastTransaction.ID
	err = s.Save()
	if err != nil {
		return err
	}
	return nil
}

// uploadTransaction uploads all nibs in the added transaction.
func (ul *Uploader) uploadTransaction(transaction *repository.Transaction) error {
	r := ul.r
	for _, nibID := range transaction.NIBIDs {
		nib, err := r.GetNIB(nibID)
		if err != nil {
			return fmt.Errorf("could not load NIB with id %s (%s)", nibID, err)
		}
		err = ul.uploadNIB(nib)
		if err != nil {
			return err
		}
	}
	return nil
}

// uploadNIBs uploads all local NIBs and content of the NIBs to
// the server.
func (ul *Uploader) uploadNIBs() error {
	r := ul.r
	nibs, err := r.GetAllNibs()
	if err != nil {
		return fmt.Errorf("unable to get NIB list (%s)", err)
	}

	for n := range nibs {
		err = ul.uploadNIB(n)
		if err != nil {
			return err
		}
	}

	return nil
}

// uploadNIB uploads a single passed NIB to the remote server.
func (ul *Uploader) uploadNIB(n *nib.NIB) error {
	r := ul.r
	client := ul.client
	objectIDs := n.AllObjectIDs()

	for _, objectID := range objectIDs {
		err := ul.uploadObject(objectID)
		if err != nil {
			return err
		}
	}
	nibReader, err := r.GetNIBReader(n.ID)
	if err != nil {
		return err
	}
	defer nibReader.Close()

	err = client.PutNIB(n.ID, nibReader)
	if err != nil {
		return fmt.Errorf("uploading nib %s failed (%s)", n.ID, err)
	}

	return nil
}

func (ul *Uploader) uploadObject(objectID string) error {
	r := ul.r
	client := ul.client

	object, err := r.GetObjectData(objectID)
	if err != nil {
		return fmt.Errorf("unable to load object %s (%s)\n", objectID, err)
	}
	defer object.Close()
	//FIXME We currently upload all objects, even multiple times
	// in some cases and even although they may already exist on
	// the server. This is not as well performing as it might be.
	err = client.PutObject(objectID, object)
	if err != nil {
		return fmt.Errorf("uploading object %s failed (%s)", objectID, err)
	}
	return nil
}
