// Code generated by protoc-gen-go.
// source: defs.proto
// DO NOT EDIT!

/*
Package odf is a generated protocol buffer package.

It is generated from these files:
	defs.proto

It has these top-level messages:
	NIB
	Revision
	Metadata
*/
package odf

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type NodeType int32

const (
	NodeType_Dir  NodeType = 0
	NodeType_File NodeType = 1
)

var NodeType_name = map[int32]string{
	0: "Dir",
	1: "File",
}
var NodeType_value = map[string]int32{
	"Dir":  0,
	"File": 1,
}

func (x NodeType) Enum() *NodeType {
	p := new(NodeType)
	*p = x
	return p
}
func (x NodeType) String() string {
	return proto.EnumName(NodeType_name, int32(x))
}
func (x *NodeType) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(NodeType_value, data, "NodeType")
	if err != nil {
		return err
	}
	*x = NodeType(value)
	return nil
}

type NIB struct {
	UUID             *string     `protobuf:"bytes,1,req" json:"UUID,omitempty"`
	Revisions        []*Revision `protobuf:"bytes,2,rep" json:"Revisions,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *NIB) Reset()         { *m = NIB{} }
func (m *NIB) String() string { return proto.CompactTextString(m) }
func (*NIB) ProtoMessage()    {}

func (m *NIB) GetUUID() string {
	if m != nil && m.UUID != nil {
		return *m.UUID
	}
	return ""
}

func (m *NIB) GetRevisions() []*Revision {
	if m != nil {
		return m.Revisions
	}
	return nil
}

type Revision struct {
	MetadataID       *string  `protobuf:"bytes,1,req" json:"MetadataID,omitempty"`
	ContentIDs       []string `protobuf:"bytes,2,rep" json:"ContentIDs,omitempty"`
	UTCTimestamp     *int64   `protobuf:"varint,3,opt" json:"UTCTimestamp,omitempty"`
	DeviceID         *string  `protobuf:"bytes,4,opt" json:"DeviceID,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *Revision) Reset()         { *m = Revision{} }
func (m *Revision) String() string { return proto.CompactTextString(m) }
func (*Revision) ProtoMessage()    {}

func (m *Revision) GetMetadataID() string {
	if m != nil && m.MetadataID != nil {
		return *m.MetadataID
	}
	return ""
}

func (m *Revision) GetContentIDs() []string {
	if m != nil {
		return m.ContentIDs
	}
	return nil
}

func (m *Revision) GetUTCTimestamp() int64 {
	if m != nil && m.UTCTimestamp != nil {
		return *m.UTCTimestamp
	}
	return 0
}

func (m *Revision) GetDeviceID() string {
	if m != nil && m.DeviceID != nil {
		return *m.DeviceID
	}
	return ""
}

type Metadata struct {
	Type             *NodeType `protobuf:"varint,1,req,enum=odf.NodeType" json:"Type,omitempty"`
	RepoRelativePath *string   `protobuf:"bytes,2,req" json:"RepoRelativePath,omitempty"`
	XXX_unrecognized []byte    `json:"-"`
}

func (m *Metadata) Reset()         { *m = Metadata{} }
func (m *Metadata) String() string { return proto.CompactTextString(m) }
func (*Metadata) ProtoMessage()    {}

func (m *Metadata) GetType() NodeType {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return NodeType_Dir
}

func (m *Metadata) GetRepoRelativePath() string {
	if m != nil && m.RepoRelativePath != nil {
		return *m.RepoRelativePath
	}
	return ""
}

func init() {
	proto.RegisterEnum("odf.NodeType", NodeType_name, NodeType_value)
}
