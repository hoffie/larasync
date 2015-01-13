package main

import (
	"fmt"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/repository"
)

// uploader handles uploads from server to client
type uploader struct {
	client *api.Client
	r      *repository.Repository
}

// pushAll ensures that the remote state is synced with the local state.
func (ul *uploader) pushAll() error {
	return ul.uploadNIBs()
}

// uploadNIBs uploads all local NIBs and content of the NIBs to
// the server.
func (ul *uploader) uploadNIBs() error {
	r := ul.r
	nibs, err := r.GetAllNibs()
	if err != nil {
		return fmt.Errorf("unable to get NIB list (%s)", err)
	}

	for nib := range nibs {
		err = ul.uploadNIB(nib)
		if err != nil {
			return err
		}
	}

	return nil
}

// uploadNIB uploads a single passed NIB to the remote server.
func (ul *uploader) uploadNIB(nib *repository.NIB) error {
	r := ul.r
	client := ul.client
	objectIDs := nib.AllObjectIDs()

	for _, objectID := range objectIDs {
		err := ul.uploadObject(objectID)
		if err != nil {
			return err
		}
	}
	nibReader, err := r.GetNIBReader(nib.ID)
	if err != nil {
		return err
	}
	defer nibReader.Close()

	//FIXME We currently assume that the server will prevent us
	// from overwriting data we are not supposed to be overwriting.
	// This will be implemented as part of #105
	err = client.PutNIB(nib.ID, nibReader)
	if err != nil {
		return fmt.Errorf("uploading nib %s failed (%s)", nib.ID, err)
	}

	return nil
}

func (ul *uploader) uploadObject(objectID string) error {
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
