package repository

import (
	"github.com/hoffie/larasync/repository/odf"
)

// Revision is part of an NIB and contains references to
// Metadata information, contents and administration data.
type Revision struct {
	MetadataID   string
	ContentIDs   []string
	UTCTimestamp int64
	DeviceID     string
}

// newRevisionFromPb returns a new Revision, pre-filled with the
// data from the given protobuf revision.
func newRevisionFromPb(pbRev *odf.Revision) *Revision {
	return &Revision{
		MetadataID:   pbRev.GetMetadataID(),
		ContentIDs:   pbRev.GetContentIDs(),
		UTCTimestamp: pbRev.GetUTCTimestamp(),
		DeviceID:     pbRev.GetDeviceID(),
	}
}

// toPb converts this Revision to a protobuf Revision.
// This is used by the encoder.
func (r *Revision) toPb() *odf.Revision {
	pb := &odf.Revision{
		MetadataID:   &r.MetadataID,
		ContentIDs:   r.ContentIDs,
		UTCTimestamp: &r.UTCTimestamp,
		DeviceID:     &r.DeviceID,
	}
	return pb
}

// AddContentID adds the given object id to the list of
// required content ids.
func (r *Revision) AddContentID(id string) error {
	r.ContentIDs = append(r.ContentIDs, id)
	return nil
}
