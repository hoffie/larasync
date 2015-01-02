package repository

import (
	"bytes"
	"errors"
	"io"

	"github.com/golang/protobuf/proto"

	"github.com/hoffie/larasync/repository/odf"
)

// NIB (Node Information Block) is a metadata object, which
// exists for every managed file or directory.
// Besides containing administration information on its own,
// it contains references to revisions.
type NIB struct {
	UUID          string
	Revisions     []*Revision
	HistoryOffset int64
}

// ReadFrom fills this NIB's data with the contents supplied by
// the binary representation available through the given reader.
func (n *NIB) ReadFrom(r io.Reader) (int64, error) {
	buf := &bytes.Buffer{}
	read, err := io.Copy(buf, r)
	if err != nil {
		return read, err
	}
	pb := &odf.NIB{}
	err = proto.Unmarshal(buf.Bytes(), pb)
	if err != nil {
		return read, err
	}
	n.UUID = pb.GetUUID()
	n.HistoryOffset = pb.GetHistoryOffset()
	if pb.Revisions != nil {
		for _, pbRev := range pb.Revisions {
			n.AppendRevision(newRevisionFromPb(pbRev))
		}
	}
	return read, nil
}

// WriteTo encodes this NIB to the supplied Writer in binary form.
// Returns the number of bytes written and an error if applicable.
func (n *NIB) WriteTo(w io.Writer) (int64, error) {
	pb := &odf.NIB{
		UUID:          &n.UUID,
		HistoryOffset: &n.HistoryOffset,
		Revisions:     make([]*odf.Revision, 0),
	}
	for _, r := range n.Revisions {
		pb.Revisions = append(pb.Revisions, r.toPb())
	}
	buf, err := proto.Marshal(pb)
	if err != nil {
		return 0, err
	}
	written, err := io.Copy(w, bytes.NewBuffer(buf))
	return written, err
}

// AppendRevision adds a new Revision to the NIB's list of
// revisions at the end.
func (n *NIB) AppendRevision(r *Revision) {
	n.Revisions = append(n.Revisions, r)
}

// LatestRevision returns the most-recently added revision.
func (n *NIB) LatestRevision() (*Revision, error) {
	l := len(n.Revisions)
	if l < 1 {
		return nil, errors.New("no revision")
	}
	return n.Revisions[l-1], nil
}

// RevisionsTotal returns the total length of all revisions.
// This is the sum of old revisions as marked by HistoryOffset plus any
// current Revisions.
func (n *NIB) RevisionsTotal() int64 {
	return int64(len(n.Revisions)) + n.HistoryOffset
}
