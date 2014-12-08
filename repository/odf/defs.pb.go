// Code generated by protoc-gen-go.
// source: defs.proto
// DO NOT EDIT!

/*
Package odf is a generated protocol buffer package.

It is generated from these files:
	defs.proto

It has these top-level messages:
	NodeInformationBlock
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
	NodeType_Directory NodeType = 0
	NodeType_File      NodeType = 1
)

var NodeType_name = map[int32]string{
	0: "Directory",
	1: "File",
}
var NodeType_value = map[string]int32{
	"Directory": 0,
	"File":      1,
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

type NodeInformationBlock struct {
	Uuid             *string     `protobuf:"bytes,1,req,name=uuid" json:"uuid,omitempty"`
	HistoryOffset    *int64      `protobuf:"varint,2,opt,name=historyOffset" json:"historyOffset,omitempty"`
	RevisionList     []*Revision `protobuf:"bytes,3,rep,name=revisionList" json:"revisionList,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *NodeInformationBlock) Reset()         { *m = NodeInformationBlock{} }
func (m *NodeInformationBlock) String() string { return proto.CompactTextString(m) }
func (*NodeInformationBlock) ProtoMessage()    {}

func (m *NodeInformationBlock) GetUuid() string {
	if m != nil && m.Uuid != nil {
		return *m.Uuid
	}
	return ""
}

func (m *NodeInformationBlock) GetHistoryOffset() int64 {
	if m != nil && m.HistoryOffset != nil {
		return *m.HistoryOffset
	}
	return 0
}

func (m *NodeInformationBlock) GetRevisionList() []*Revision {
	if m != nil {
		return m.RevisionList
	}
	return nil
}

type Revision struct {
	MetadataID       *string `protobuf:"bytes,1,req,name=metadataID" json:"metadataID,omitempty"`
	ContentIDList    *string `protobuf:"bytes,2,req,name=contentIDList" json:"contentIDList,omitempty"`
	UtcTimestamp     *int64  `protobuf:"varint,3,opt,name=utcTimestamp" json:"utcTimestamp,omitempty"`
	DeviceID         *string `protobuf:"bytes,4,opt,name=deviceID" json:"deviceID,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
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

func (m *Revision) GetContentIDList() string {
	if m != nil && m.ContentIDList != nil {
		return *m.ContentIDList
	}
	return ""
}

func (m *Revision) GetUtcTimestamp() int64 {
	if m != nil && m.UtcTimestamp != nil {
		return *m.UtcTimestamp
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
	NodeType         *NodeType `protobuf:"varint,1,req,name=nodeType,enum=odf.NodeType" json:"nodeType,omitempty"`
	RepoRelativePath *string   `protobuf:"bytes,2,req,name=repoRelativePath" json:"repoRelativePath,omitempty"`
	XXX_unrecognized []byte    `json:"-"`
}

func (m *Metadata) Reset()         { *m = Metadata{} }
func (m *Metadata) String() string { return proto.CompactTextString(m) }
func (*Metadata) ProtoMessage()    {}

func (m *Metadata) GetNodeType() NodeType {
	if m != nil && m.NodeType != nil {
		return *m.NodeType
	}
	return NodeType_Directory
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
