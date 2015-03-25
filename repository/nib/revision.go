package nib

import (
	"reflect"

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

// HasSameContent returns whether this revision's metadata and
// content ids match the ids of the provided other revision instance.
func (r *Revision) HasSameContent(other *Revision) bool {
	if r.MetadataID != other.MetadataID {
		return false
	}
	if !reflect.DeepEqual(r.ContentIDs, other.ContentIDs) {
		return false
	}
	return true
}

// IsDelete returns if the revision marks the item as being deleted.
func (r *Revision) IsDeletion() bool {
	return len(r.ContentIDs) == 0
}

// Clone returns an exact copy of the given revision.
func (r *Revision) Clone() *Revision {
	return &Revision{
		MetadataID:   r.MetadataID,
		ContentIDs:   r.cloneContentIDs(),
		UTCTimestamp: r.UTCTimestamp,
		DeviceID:     r.DeviceID,
	}
}

// cloneContentIDs returns a new array populated with the current list
// of contentIDs.
func (r *Revision) cloneContentIDs() []string {
	return append([]string{}, r.ContentIDs...)
}
