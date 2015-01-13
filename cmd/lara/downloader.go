package main

import (
	"bytes"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/repository"
)

// downloader handles downloads from server to client
type downloader struct {
	client *api.Client
	r      *repository.Repository
}

// getAll ensures that the local state matches the remote state.
func (dl *downloader) getAll() error {
	err := dl.getNIBs()
	if err != nil {
		return err
	}
	err = dl.getMissingObjects()
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
		err := dl.r.AddNIBContent(bytes.NewReader(nibBytes))
		if err != nil {
			return err
		}
	}
	return nil
}

// getMissingObjects parses all NIBs and retrieves all
// missing objects.
func (dl *downloader) getMissingObjects() error {
	nibs, err := dl.r.GetAllNibs()
	if err != nil {
		return err
	}
	for nib := range nibs {
		objectIDs := nib.AllObjectIDs()
		for _, objectID := range objectIDs {
			if dl.r.HasObject(objectID) {
				continue
			}
			err = dl.getObject(objectID)
			if err != nil {
				return err
			}
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