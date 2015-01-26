package client

import (
	"bytes"

	"github.com/hoffie/larasync/repository"
	"github.com/hoffie/larasync/repository/nib"
)

// Downloader returns the downloader configured for the client
// and the passed ClientRepository.
func (c *Client) Downloader(r *repository.ClientRepository) *Downloader {
	return &Downloader{
		client: c,
		r:      r,
	}
}

// Downloader handles downloads from server to client
type Downloader struct {
	client *Client
	r      *repository.ClientRepository
}

// GetAll ensures that the local state matches the remote state.
func (dl *Downloader) GetAll() error {
	err := dl.getNIBs()
	if err != nil {
		return err
	}
	return nil
}

// GetDelta ensures that all data from the last synced transaction
// is in sync with the local one.
func (dl *Downloader) GetDelta() error {
	repository := dl.r
	stateConfig, err := repository.StateConfig()
	if err != nil {
		return err
	}
	defaultServer := stateConfig.DefaultServer
	remoteTransactionID := defaultServer.RemoteTransactionID
	if remoteTransactionID == 0 {
		err = dl.getNIBs()
	} else {
		err = dl.getFromServerTransactionID(defaultServer.RemoteTransactionID)
	}
	return err
}

// getFromServerTransactionID syncs all data from the given server transaction
// ID.
func (dl *Downloader) getFromServerTransactionID(transactionID int64) error {
	nibResponse, err := dl.client.GetNIBsFromTransactionID(transactionID)
	if err != nil {
		return err
	}
	return dl.processNIBResponse(nibResponse)
}

// getNIBs downloads all NIBs and stores them in the repository
func (dl *Downloader) getNIBs() error {
	nibResponse, err := dl.client.GetNIBs()
	if err != nil {
		return err
	}
	return dl.processNIBResponse(nibResponse)
}

// processNIBResponse synchronizes the given NIBResponse to the local client state.
func (dl *Downloader) processNIBResponse(response *NIBGetResponse) error {
	err := dl.processNIBBytes(response.NIBData)
	if err != nil {
		return err
	}
	serverTransaction := response.ServerTransactionID
	stateConfig, err := dl.r.StateConfig()
	if err != nil {
		return err
	}
	stateConfig.DefaultServer.RemoteTransactionID = serverTransaction
	return stateConfig.Save()
}

// processNibBytes parses a channel and adds the NIBs being represented by each
// passed byte array.
func (dl *Downloader) processNIBBytes(nibBytesIterator <-chan []byte) error {
	for nibBytes := range nibBytesIterator {
		// FIXME: overwrite checking!
		n, err := dl.r.VerifyAndParseNIBBytes(nibBytes)
		if err != nil {
			return err
		}
		err = dl.fetchMissingData(n)
		if err != nil {
			return err
		}

		err = dl.r.AddNIBContent(bytes.NewReader(nibBytes))
		if err != nil {
			return err
		}
	}
	return nil
}

// fetchMissingData loads missing objects in the passed NIB.
func (dl *Downloader) fetchMissingData(n *nib.NIB) error {
	objectIDs := n.AllObjectIDs()
	for _, objectID := range objectIDs {
		if dl.r.HasObject(objectID) {
			continue
		}
		err := dl.getObject(objectID)
		if err != nil {
			return err
		}
	}
	return nil
}

// getObject downloads the named object
func (dl *Downloader) getObject(objectID string) error {
	resp, err := dl.client.GetObject(objectID)
	if err != nil {
		return err
	}
	err = dl.r.AddObject(objectID, resp)
	return err
}
