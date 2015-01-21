package main

import (
	"fmt"

	"github.com/hoffie/larasync/api/client"
	"github.com/hoffie/larasync/repository"
	"github.com/hoffie/larasync/repository/nib"
)

// uploader handles uploads from server to client
type uploader struct {
	client *client.Client
	r      *repository.ClientRepository
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

	for n := range nibs {
		err = ul.uploadNIB(n)
		if err != nil {
			return err
		}
	}

	return nil
}

// uploadNIB uploads a single passed NIB to the remote server.
func (ul *uploader) uploadNIB(n *nib.NIB) error {
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
