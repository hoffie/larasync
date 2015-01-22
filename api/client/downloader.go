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

// getNIBs downloads all NIBs and stores them in the repository
func (dl *Downloader) getNIBs() error {
	nibBytesIterator, err := dl.client.GetNIBs()
	if err != nil {
		return err
	}
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
