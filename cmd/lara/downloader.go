package main

import (
	"bytes"

	"github.com/hoffie/larasync/api/client"
	"github.com/hoffie/larasync/repository"
	"github.com/hoffie/larasync/repository/nib"
)

// downloader handles downloads from server to client
type downloader struct {
	client *client.Client
	r      *repository.ClientRepository
}

// getAll ensures that the local state matches the remote state.
func (dl *downloader) getAll() error {
	err := dl.getNIBs()
	if err != nil {
		return err
	}
	return nil
}

// getNIBs downloads all NIBs and stores them in the repository
func (dl *downloader) getNIBs() error {
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
func (dl *downloader) fetchMissingData(n *nib.NIB) error {
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
func (dl *downloader) getObject(objectID string) error {
	resp, err := dl.client.GetObject(objectID)
	if err != nil {
		return err
	}
	err = dl.r.AddObject(objectID, resp)
	return err
}
