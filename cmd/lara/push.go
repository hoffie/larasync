package main

import (
	"fmt"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/repository"
)

// pushAction implements "lara push"
func (d *Dispatcher) pushAction() int {
	if len(d.flags.Args()) != 0 {
		fmt.Fprint(d.stderr, "Error: this command takes no arguments\n")
		return 1
	}
	root, err := d.getRootFromWd()
	if err != nil {
		return 1
	}
	r := repository.New(root)

	client, err := clientFor(r)
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: %s\n",
			err)
	}

	//FIXME:
	nibs, err := r.GetAllNibs()
	if err != nil {
		fmt.Fprintf(d.stderr, "Error: unable to get NIB list (%s)\n", err)
		return 1
	}
	for nib := range nibs {
		objectIDs := nib.AllObjectIDs()
		for _, objectID := range objectIDs {
			object, err := r.GetObjectData(objectID)
			if err != nil {
				fmt.Fprintf(d.stderr, "Error: unable to load object %s (%s)\n",
					objectID, err)
				return 1
			}
			defer object.Close()
			//FIXME We currently upload all objects, even multiple times
			// in some cases and even although they may already exist on
			// the server. This is not as well performing as it might be.
			err = client.PutObject(objectID, object)
			if err != nil {
				fmt.Fprintf(d.stderr, "Error: uploading object %s failed (%s)\n",
					objectID, err)
				return 1
			}
		}
		nibReader, err := r.GetNIBReader(nib.ID)
		if err != nil {
			fmt.Fprintf(d.stderr, "Error: unable to load nib %s (%s)\n",
				nib.ID, err)
			return 1
		}
		//FIXME We currently assume that the server will prevent us
		// from overwriting data we are not supposed to be overwriting.
		// This will be implemented as part of #105
		err = client.PutNIB(nib.ID, nibReader)
		if err != nil {
			fmt.Fprintf(d.stderr, "Error: uploading nib %s failed (%s)\n",
				nib.ID, err)
			return 1
		}
	}
	return 0
}
