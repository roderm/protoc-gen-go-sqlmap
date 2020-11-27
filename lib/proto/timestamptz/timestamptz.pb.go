// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: lib/proto/timestamptz/timestamptz.proto

package timestamptz

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type TimestampTZ struct {
	Seconds              int64    `protobuf:"varint,1,opt,name=seconds,proto3" json:"seconds,omitempty"`
	Nanos                int32    `protobuf:"varint,2,opt,name=nanos,proto3" json:"nanos,omitempty"`
	Zone                 string   `protobuf:"bytes,3,opt,name=Zone,proto3" json:"Zone,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TimestampTZ) Reset()         { *m = TimestampTZ{} }
func (m *TimestampTZ) String() string { return proto.CompactTextString(m) }
func (*TimestampTZ) ProtoMessage()    {}
func (*TimestampTZ) Descriptor() ([]byte, []int) {
	return fileDescriptor_ecd459969fd81cb2, []int{0}
}
func (m *TimestampTZ) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TimestampTZ.Unmarshal(m, b)
}
func (m *TimestampTZ) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TimestampTZ.Marshal(b, m, deterministic)
}
func (m *TimestampTZ) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TimestampTZ.Merge(m, src)
}
func (m *TimestampTZ) XXX_Size() int {
	return xxx_messageInfo_TimestampTZ.Size(m)
}
func (m *TimestampTZ) XXX_DiscardUnknown() {
	xxx_messageInfo_TimestampTZ.DiscardUnknown(m)
}

var xxx_messageInfo_TimestampTZ proto.InternalMessageInfo

func (m *TimestampTZ) GetSeconds() int64 {
	if m != nil {
		return m.Seconds
	}
	return 0
}

func (m *TimestampTZ) GetNanos() int32 {
	if m != nil {
		return m.Nanos
	}
	return 0
}

func (m *TimestampTZ) GetZone() string {
	if m != nil {
		return m.Zone
	}
	return ""
}

func init() {
	proto.RegisterType((*TimestampTZ)(nil), "timestamptz.TimestampTZ")
}

func init() {
	proto.RegisterFile("lib/proto/timestamptz/timestamptz.proto", fileDescriptor_ecd459969fd81cb2)
}

var fileDescriptor_ecd459969fd81cb2 = []byte{
	// 165 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0xcf, 0xc9, 0x4c, 0xd2,
	0x2f, 0x28, 0xca, 0x2f, 0xc9, 0xd7, 0x2f, 0xc9, 0xcc, 0x4d, 0x2d, 0x2e, 0x49, 0xcc, 0x2d, 0x28,
	0xa9, 0x42, 0x66, 0xeb, 0x81, 0x65, 0x85, 0xb8, 0x91, 0x84, 0x94, 0x02, 0xb9, 0xb8, 0x43, 0x60,
	0xdc, 0x90, 0x28, 0x21, 0x09, 0x2e, 0xf6, 0xe2, 0xd4, 0xe4, 0xfc, 0xbc, 0x94, 0x62, 0x09, 0x46,
	0x05, 0x46, 0x0d, 0xe6, 0x20, 0x18, 0x57, 0x48, 0x84, 0x8b, 0x35, 0x2f, 0x31, 0x2f, 0xbf, 0x58,
	0x82, 0x49, 0x81, 0x51, 0x83, 0x35, 0x08, 0xc2, 0x11, 0x12, 0xe2, 0x62, 0x89, 0xca, 0xcf, 0x4b,
	0x95, 0x60, 0x56, 0x60, 0xd4, 0xe0, 0x0c, 0x02, 0xb3, 0x9d, 0xec, 0xa2, 0x6c, 0xd2, 0x33, 0x4b,
	0x32, 0x4a, 0x93, 0xf4, 0x92, 0xf3, 0x73, 0xf5, 0x8b, 0xf2, 0x53, 0x52, 0x8b, 0x72, 0x21, 0x0e,
	0x4b, 0xd6, 0x4d, 0x4f, 0xcd, 0xd3, 0x4d, 0xcf, 0xd7, 0x2d, 0x2e, 0xcc, 0xc9, 0x4d, 0x2c, 0xd0,
	0xc7, 0xea, 0xe2, 0x24, 0x36, 0xb0, 0x90, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0xce, 0xf6, 0x08,
	0x09, 0xd1, 0x00, 0x00, 0x00,
}
