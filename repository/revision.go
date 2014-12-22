package repository

import (
	"github.com/hoffie/larasync/repository/odf"
)

// Revision is part of an NIB and contains references to
// Metadata information, contents and administration data.
type Revision struct {
	MetadataID string
}

// newRevisionFromPb returns a new Revision, pre-filled with the
// data from the given protobuf revision.
func newRevisionFromPb(pbRev *odf.Revision) *Revision {
	return &Revision{
		MetadataID: pbRev.GetMetadataID(),
	}
}

// toPb converts this Revision to a protobuf Revision.
// This is used by the encoder.
func (r *Revision) toPb() *odf.Revision {
	pb := &odf.Revision{
		MetadataID: &r.MetadataID,
	}
	return pb
}
