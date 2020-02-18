// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: node.proto

package pb

import (
	fmt "fmt"
	math "math"
	time "time"

	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/golang/protobuf/ptypes/timestamp"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// NodeType is an enum of possible node types
type NodeType int32

const (
	NodeType_INVALID   NodeType = 0
	NodeType_SATELLITE NodeType = 1
	NodeType_STORAGE   NodeType = 2
	NodeType_UPLINK    NodeType = 3
	NodeType_BOOTSTRAP NodeType = 4 // Deprecated: Do not use.
)

var NodeType_name = map[int32]string{
	0: "INVALID",
	1: "SATELLITE",
	2: "STORAGE",
	3: "UPLINK",
	4: "BOOTSTRAP",
}

var NodeType_value = map[string]int32{
	"INVALID":   0,
	"SATELLITE": 1,
	"STORAGE":   2,
	"UPLINK":    3,
	"BOOTSTRAP": 4,
}

func (x NodeType) String() string {
	return proto.EnumName(NodeType_name, int32(x))
}

func (NodeType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{0}
}

// NodeTransport is an enum of possible transports for the overlay network
type NodeTransport int32

const (
	NodeTransport_TCP_TLS_GRPC NodeTransport = 0
)

var NodeTransport_name = map[int32]string{
	0: "TCP_TLS_GRPC",
}

var NodeTransport_value = map[string]int32{
	"TCP_TLS_GRPC": 0,
}

func (x NodeTransport) String() string {
	return proto.EnumName(NodeTransport_name, int32(x))
}

func (NodeTransport) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{1}
}

// TODO move statdb.Update() stuff out of here
// Node represents a node in the overlay network
// Node is info for a updating a single storagenode, used in the Update rpc calls
type Node struct {
	Id                   NodeID       `protobuf:"bytes,1,opt,name=id,proto3,customtype=NodeID" json:"id"`
	Address              *NodeAddress `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	LastIp               string       `protobuf:"bytes,14,opt,name=last_ip,json=lastIp,proto3" json:"last_ip,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *Node) Reset()         { *m = Node{} }
func (m *Node) String() string { return proto.CompactTextString(m) }
func (*Node) ProtoMessage()    {}
func (*Node) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{0}
}
func (m *Node) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Node.Unmarshal(m, b)
}
func (m *Node) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Node.Marshal(b, m, deterministic)
}
func (m *Node) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Node.Merge(m, src)
}
func (m *Node) XXX_Size() int {
	return xxx_messageInfo_Node.Size(m)
}
func (m *Node) XXX_DiscardUnknown() {
	xxx_messageInfo_Node.DiscardUnknown(m)
}

var xxx_messageInfo_Node proto.InternalMessageInfo

func (m *Node) GetAddress() *NodeAddress {
	if m != nil {
		return m.Address
	}
	return nil
}

func (m *Node) GetLastIp() string {
	if m != nil {
		return m.LastIp
	}
	return ""
}

// NodeAddress contains the information needed to communicate with a node on the network
type NodeAddress struct {
	Transport            NodeTransport `protobuf:"varint,1,opt,name=transport,proto3,enum=node.NodeTransport" json:"transport,omitempty"`
	Address              string        `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *NodeAddress) Reset()         { *m = NodeAddress{} }
func (m *NodeAddress) String() string { return proto.CompactTextString(m) }
func (*NodeAddress) ProtoMessage()    {}
func (*NodeAddress) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{1}
}
func (m *NodeAddress) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeAddress.Unmarshal(m, b)
}
func (m *NodeAddress) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeAddress.Marshal(b, m, deterministic)
}
func (m *NodeAddress) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeAddress.Merge(m, src)
}
func (m *NodeAddress) XXX_Size() int {
	return xxx_messageInfo_NodeAddress.Size(m)
}
func (m *NodeAddress) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeAddress.DiscardUnknown(m)
}

var xxx_messageInfo_NodeAddress proto.InternalMessageInfo

func (m *NodeAddress) GetTransport() NodeTransport {
	if m != nil {
		return m.Transport
	}
	return NodeTransport_TCP_TLS_GRPC
}

func (m *NodeAddress) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

// NodeOperator contains info about the storage node operator
type NodeOperator struct {
	Email                string   `protobuf:"bytes,1,opt,name=email,proto3" json:"email,omitempty"`
	Wallet               string   `protobuf:"bytes,2,opt,name=wallet,proto3" json:"wallet,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NodeOperator) Reset()         { *m = NodeOperator{} }
func (m *NodeOperator) String() string { return proto.CompactTextString(m) }
func (*NodeOperator) ProtoMessage()    {}
func (*NodeOperator) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{2}
}
func (m *NodeOperator) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeOperator.Unmarshal(m, b)
}
func (m *NodeOperator) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeOperator.Marshal(b, m, deterministic)
}
func (m *NodeOperator) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeOperator.Merge(m, src)
}
func (m *NodeOperator) XXX_Size() int {
	return xxx_messageInfo_NodeOperator.Size(m)
}
func (m *NodeOperator) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeOperator.DiscardUnknown(m)
}

var xxx_messageInfo_NodeOperator proto.InternalMessageInfo

func (m *NodeOperator) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *NodeOperator) GetWallet() string {
	if m != nil {
		return m.Wallet
	}
	return ""
}

// NodeCapacity contains all relevant data about a nodes ability to store data
type NodeCapacity struct {
	FreeBandwidth        int64    `protobuf:"varint,1,opt,name=free_bandwidth,json=freeBandwidth,proto3" json:"free_bandwidth,omitempty"`
	FreeDisk             int64    `protobuf:"varint,2,opt,name=free_disk,json=freeDisk,proto3" json:"free_disk,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NodeCapacity) Reset()         { *m = NodeCapacity{} }
func (m *NodeCapacity) String() string { return proto.CompactTextString(m) }
func (*NodeCapacity) ProtoMessage()    {}
func (*NodeCapacity) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{3}
}
func (m *NodeCapacity) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeCapacity.Unmarshal(m, b)
}
func (m *NodeCapacity) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeCapacity.Marshal(b, m, deterministic)
}
func (m *NodeCapacity) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeCapacity.Merge(m, src)
}
func (m *NodeCapacity) XXX_Size() int {
	return xxx_messageInfo_NodeCapacity.Size(m)
}
func (m *NodeCapacity) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeCapacity.DiscardUnknown(m)
}

var xxx_messageInfo_NodeCapacity proto.InternalMessageInfo

func (m *NodeCapacity) GetFreeBandwidth() int64 {
	if m != nil {
		return m.FreeBandwidth
	}
	return 0
}

func (m *NodeCapacity) GetFreeDisk() int64 {
	if m != nil {
		return m.FreeDisk
	}
	return 0
}

// Deprecated: use NodeOperator instead
type NodeMetadata struct {
	Email                string   `protobuf:"bytes,1,opt,name=email,proto3" json:"email,omitempty"`
	Wallet               string   `protobuf:"bytes,2,opt,name=wallet,proto3" json:"wallet,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NodeMetadata) Reset()         { *m = NodeMetadata{} }
func (m *NodeMetadata) String() string { return proto.CompactTextString(m) }
func (*NodeMetadata) ProtoMessage()    {}
func (*NodeMetadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{4}
}
func (m *NodeMetadata) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeMetadata.Unmarshal(m, b)
}
func (m *NodeMetadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeMetadata.Marshal(b, m, deterministic)
}
func (m *NodeMetadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeMetadata.Merge(m, src)
}
func (m *NodeMetadata) XXX_Size() int {
	return xxx_messageInfo_NodeMetadata.Size(m)
}
func (m *NodeMetadata) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeMetadata.DiscardUnknown(m)
}

var xxx_messageInfo_NodeMetadata proto.InternalMessageInfo

func (m *NodeMetadata) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *NodeMetadata) GetWallet() string {
	if m != nil {
		return m.Wallet
	}
	return ""
}

// Deprecated: use NodeCapacity instead
type NodeRestrictions struct {
	FreeBandwidth        int64    `protobuf:"varint,1,opt,name=free_bandwidth,json=freeBandwidth,proto3" json:"free_bandwidth,omitempty"`
	FreeDisk             int64    `protobuf:"varint,2,opt,name=free_disk,json=freeDisk,proto3" json:"free_disk,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NodeRestrictions) Reset()         { *m = NodeRestrictions{} }
func (m *NodeRestrictions) String() string { return proto.CompactTextString(m) }
func (*NodeRestrictions) ProtoMessage()    {}
func (*NodeRestrictions) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{5}
}
func (m *NodeRestrictions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeRestrictions.Unmarshal(m, b)
}
func (m *NodeRestrictions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeRestrictions.Marshal(b, m, deterministic)
}
func (m *NodeRestrictions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeRestrictions.Merge(m, src)
}
func (m *NodeRestrictions) XXX_Size() int {
	return xxx_messageInfo_NodeRestrictions.Size(m)
}
func (m *NodeRestrictions) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeRestrictions.DiscardUnknown(m)
}

var xxx_messageInfo_NodeRestrictions proto.InternalMessageInfo

func (m *NodeRestrictions) GetFreeBandwidth() int64 {
	if m != nil {
		return m.FreeBandwidth
	}
	return 0
}

func (m *NodeRestrictions) GetFreeDisk() int64 {
	if m != nil {
		return m.FreeDisk
	}
	return 0
}

// NodeVersion contains
type NodeVersion struct {
	Version              string    `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	CommitHash           string    `protobuf:"bytes,2,opt,name=commit_hash,json=commitHash,proto3" json:"commit_hash,omitempty"`
	Timestamp            time.Time `protobuf:"bytes,3,opt,name=timestamp,proto3,stdtime" json:"timestamp"`
	Release              bool      `protobuf:"varint,4,opt,name=release,proto3" json:"release,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *NodeVersion) Reset()         { *m = NodeVersion{} }
func (m *NodeVersion) String() string { return proto.CompactTextString(m) }
func (*NodeVersion) ProtoMessage()    {}
func (*NodeVersion) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{6}
}
func (m *NodeVersion) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeVersion.Unmarshal(m, b)
}
func (m *NodeVersion) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeVersion.Marshal(b, m, deterministic)
}
func (m *NodeVersion) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeVersion.Merge(m, src)
}
func (m *NodeVersion) XXX_Size() int {
	return xxx_messageInfo_NodeVersion.Size(m)
}
func (m *NodeVersion) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeVersion.DiscardUnknown(m)
}

var xxx_messageInfo_NodeVersion proto.InternalMessageInfo

func (m *NodeVersion) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *NodeVersion) GetCommitHash() string {
	if m != nil {
		return m.CommitHash
	}
	return ""
}

func (m *NodeVersion) GetTimestamp() time.Time {
	if m != nil {
		return m.Timestamp
	}
	return time.Time{}
}

func (m *NodeVersion) GetRelease() bool {
	if m != nil {
		return m.Release
	}
	return false
}

func init() {
	proto.RegisterEnum("node.NodeType", NodeType_name, NodeType_value)
	proto.RegisterEnum("node.NodeTransport", NodeTransport_name, NodeTransport_value)
	proto.RegisterType((*Node)(nil), "node.Node")
	proto.RegisterType((*NodeAddress)(nil), "node.NodeAddress")
	proto.RegisterType((*NodeOperator)(nil), "node.NodeOperator")
	proto.RegisterType((*NodeCapacity)(nil), "node.NodeCapacity")
	proto.RegisterType((*NodeMetadata)(nil), "node.NodeMetadata")
	proto.RegisterType((*NodeRestrictions)(nil), "node.NodeRestrictions")
	proto.RegisterType((*NodeVersion)(nil), "node.NodeVersion")
}

func init() { proto.RegisterFile("node.proto", fileDescriptor_0c843d59d2d938e7) }

var fileDescriptor_0c843d59d2d938e7 = []byte{
	// 591 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x53, 0xdd, 0x6e, 0xd3, 0x30,
	0x14, 0x5e, 0xda, 0xac, 0x4d, 0x4e, 0x7f, 0x94, 0x99, 0x09, 0xaa, 0x22, 0xd1, 0x52, 0x09, 0xa9,
	0x1a, 0x52, 0x27, 0xc6, 0x2d, 0x37, 0xed, 0x36, 0x8d, 0x42, 0x59, 0x2b, 0x37, 0xec, 0x62, 0x37,
	0x91, 0xdb, 0x78, 0xad, 0xb5, 0x34, 0xb6, 0x6c, 0x87, 0xa9, 0x6f, 0xc1, 0x53, 0xf0, 0x2c, 0x3c,
	0x03, 0x17, 0xe3, 0x4d, 0x10, 0x72, 0x7e, 0xb6, 0xf5, 0x12, 0x89, 0xbb, 0x7c, 0xdf, 0xf9, 0x8e,
	0xfd, 0xe5, 0x3b, 0xc7, 0x00, 0x31, 0x0f, 0xe9, 0x40, 0x48, 0xae, 0x39, 0xb2, 0xcd, 0x77, 0x1b,
	0x56, 0x7c, 0xc5, 0x33, 0xa6, 0xdd, 0x59, 0x71, 0xbe, 0x8a, 0xe8, 0x71, 0x8a, 0x16, 0xc9, 0xcd,
	0xb1, 0x66, 0x1b, 0xaa, 0x34, 0xd9, 0x88, 0x4c, 0xd0, 0xfb, 0x63, 0x81, 0x7d, 0xc9, 0x43, 0x8a,
	0x5e, 0x41, 0x89, 0x85, 0x2d, 0xab, 0x6b, 0xf5, 0xeb, 0xa3, 0xe6, 0xcf, 0xfb, 0xce, 0xde, 0xaf,
	0xfb, 0x4e, 0xc5, 0x54, 0xc6, 0x67, 0xb8, 0xc4, 0x42, 0xf4, 0x16, 0xaa, 0x24, 0x0c, 0x25, 0x55,
	0xaa, 0x55, 0xea, 0x5a, 0xfd, 0xda, 0xc9, 0xc1, 0x20, 0xbd, 0xd9, 0x48, 0x86, 0x59, 0x01, 0x17,
	0x0a, 0xf4, 0x02, 0xaa, 0x11, 0x51, 0x3a, 0x60, 0xa2, 0xd5, 0xec, 0x5a, 0x7d, 0x17, 0x57, 0x0c,
	0x1c, 0x8b, 0x4f, 0xb6, 0x53, 0xf6, 0x9a, 0xd8, 0xd6, 0x5b, 0x41, 0x71, 0x5d, 0x52, 0xa5, 0x25,
	0x5b, 0x6a, 0xc6, 0x63, 0x85, 0x41, 0x52, 0x91, 0x68, 0x62, 0x00, 0x76, 0x36, 0x54, 0x93, 0x90,
	0x68, 0x82, 0xeb, 0x11, 0xd1, 0x34, 0x5e, 0x6e, 0x83, 0x88, 0x29, 0x8d, 0x1b, 0x24, 0x09, 0x99,
	0x0e, 0x54, 0xb2, 0x5c, 0x9a, 0xeb, 0xf6, 0x99, 0x0a, 0x12, 0x81, 0x9b, 0x89, 0x08, 0x89, 0xa6,
	0x41, 0x2e, 0xc5, 0x87, 0x39, 0xde, 0x15, 0x37, 0x72, 0x36, 0x11, 0x26, 0x02, 0x5c, 0xfd, 0x46,
	0xa5, 0x62, 0x3c, 0xee, 0x5d, 0x43, 0xed, 0xc9, 0x2f, 0xa0, 0x77, 0xe0, 0x6a, 0x49, 0x62, 0x25,
	0xb8, 0xd4, 0x69, 0x1a, 0xcd, 0x93, 0x67, 0x8f, 0x3f, 0xea, 0x17, 0x25, 0xfc, 0xa8, 0x42, 0xad,
	0xdd, 0x64, 0xdc, 0x87, 0x18, 0x7a, 0x1f, 0xa0, 0x6e, 0xba, 0xa6, 0x82, 0x4a, 0xa2, 0xb9, 0x44,
	0x87, 0xb0, 0x4f, 0x37, 0x84, 0x45, 0xe9, 0xc1, 0x2e, 0xce, 0x00, 0x7a, 0x0e, 0x95, 0x3b, 0x12,
	0x45, 0x54, 0xe7, 0xed, 0x39, 0xea, 0xe1, 0xac, 0xfb, 0x94, 0x08, 0xb2, 0x64, 0x7a, 0x8b, 0xde,
	0x40, 0xf3, 0x46, 0x52, 0x1a, 0x2c, 0x48, 0x1c, 0xde, 0xb1, 0x50, 0xaf, 0xd3, 0x63, 0xca, 0xb8,
	0x61, 0xd8, 0x51, 0x41, 0xa2, 0x97, 0xe0, 0xa6, 0xb2, 0x90, 0xa9, 0xdb, 0xf4, 0xc4, 0x32, 0x76,
	0x0c, 0x71, 0xc6, 0xd4, 0x6d, 0xe1, 0xe8, 0x4b, 0x9e, 0xef, 0x3f, 0x3a, 0xba, 0x02, 0xcf, 0x74,
	0xe3, 0x27, 0x73, 0xfb, 0x2f, 0xae, 0x7e, 0x58, 0xd9, 0x10, 0xae, 0xb2, 0x99, 0x98, 0x44, 0xf3,
	0xf1, 0xe4, 0xbe, 0x0a, 0x88, 0x3a, 0x50, 0x5b, 0xf2, 0xcd, 0x86, 0xe9, 0x60, 0x4d, 0xd4, 0x3a,
	0xb7, 0x07, 0x19, 0xf5, 0x91, 0xa8, 0x35, 0x1a, 0x81, 0xfb, 0xb0, 0xe2, 0xad, 0x72, 0xba, 0xa8,
	0xed, 0x41, 0xf6, 0x08, 0x06, 0xc5, 0x23, 0x18, 0xf8, 0x85, 0x62, 0xe4, 0x98, 0x4d, 0xff, 0xfe,
	0xbb, 0x63, 0xe1, 0xc7, 0x36, 0x73, 0xbd, 0xa4, 0x11, 0x25, 0x8a, 0xb6, 0xec, 0xae, 0xd5, 0x77,
	0x70, 0x01, 0x8f, 0x30, 0x38, 0xe9, 0x1a, 0x6c, 0x05, 0x45, 0x35, 0xa8, 0x8e, 0x2f, 0xaf, 0x86,
	0x93, 0xf1, 0x99, 0xb7, 0x87, 0x1a, 0xe0, 0xce, 0x87, 0xfe, 0xf9, 0x64, 0x32, 0xf6, 0xcf, 0x3d,
	0xcb, 0xd4, 0xe6, 0xfe, 0x14, 0x0f, 0x2f, 0xce, 0xbd, 0x12, 0x02, 0xa8, 0x7c, 0x9d, 0x4d, 0xc6,
	0x97, 0x9f, 0xbd, 0x32, 0x3a, 0x00, 0x77, 0x34, 0x9d, 0xfa, 0x73, 0x1f, 0x0f, 0x67, 0x9e, 0xdd,
	0x2e, 0x39, 0xd6, 0xd1, 0x6b, 0x68, 0xec, 0xac, 0x16, 0xf2, 0xa0, 0xee, 0x9f, 0xce, 0x02, 0x7f,
	0x32, 0x0f, 0x2e, 0xf0, 0xec, 0xd4, 0xdb, 0x1b, 0xd9, 0xd7, 0x25, 0xb1, 0x58, 0x54, 0x52, 0xff,
	0xef, 0xff, 0x06, 0x00, 0x00, 0xff, 0xff, 0xf0, 0xab, 0x88, 0x34, 0xf2, 0x03, 0x00, 0x00,
}
