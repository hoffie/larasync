package repository

import (
	"bytes"
	"io"

	"github.com/golang/protobuf/proto"

	"github.com/hoffie/larasync/repository/odf"
)

const (
	MetadataTypeFile = iota
	MetadataTypeDir
)

// Metadata describes information about a node such as its type and name.
type Metadata struct {
	Type             int32
	RepoRelativePath string
}

// newMetadataFromPb returns a new Metadata object, pre-filled with the
// data from the given protobuf object.
func newMetadataFromPb(pbRev *odf.Metadata) *Metadata {
	return &Metadata{}
}

// toPb converts this Metadata object to a protobuf object.
// This is used by the encoder.
func (m *Metadata) toPb() *odf.Metadata {
	pb := &odf.Metadata{}
	return pb
}

// WriteTo encodes this Metadata object to the supplied Writer in binary
// form.
// Returns the number of bytes written and an error if applicable.
func (m *Metadata) WriteTo(w io.Writer) (int64, error) {
	t := odf.NodeType_File
	if m.Type == MetadataTypeDir {
		t = odf.NodeType_Dir
	}
	pb := &odf.Metadata{
		Type:             &t,
		RepoRelativePath: &m.RepoRelativePath,
	}
	buf, err := proto.Marshal(pb)
	if err != nil {
		return 0, err
	}
	written, err := io.Copy(w, bytes.NewBuffer(buf))
	return written, err
}

// ReadFrom fills this Metadata's data with the contents supplied by
// the binary representation available through the given reader.
func (m *Metadata) ReadFrom(r io.Reader) (int64, error) {
	buf := &bytes.Buffer{}
	read, err := io.Copy(buf, r)
	if err != nil {
		return read, err
	}
	pb := &odf.Metadata{}
	err = proto.Unmarshal(buf.Bytes(), pb)
	if err != nil {
		return read, err
	}
	m.RepoRelativePath = pb.GetRepoRelativePath()
	if pb.GetType() == odf.NodeType_Dir {
		m.Type = MetadataTypeDir
	} else {
		m.Type = MetadataTypeFile
	}
	return read, nil
}
