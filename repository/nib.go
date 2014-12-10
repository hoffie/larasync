package repository

import (
	"io"
	"bytes"

	"github.com/golang/protobuf/proto"

	"github.com/hoffie/larasync/repository/odf"
)

type NIB struct {
	UUID string
}

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
	n.UUID = *pb.UUID
	return read, nil
}

func (n *NIB) WriteTo(w io.Writer) (int64, error) {
	pb := &odf.NIB{
		UUID: &n.UUID,
	}
	buf, err := proto.Marshal(pb)
	if err != nil {
		return 0, err
	}
	written, err := io.Copy(w, bytes.NewBuffer(buf))
	return written, err
}
