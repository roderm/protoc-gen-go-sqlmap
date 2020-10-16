// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sqlgen/eyevip.proto

package sqlgen

import (
	fmt "fmt"
	descriptor "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	proto "github.com/golang/protobuf/proto"
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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

var E_Dbcol = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.FieldOptions)(nil),
	ExtensionType: (*string)(nil),
	Field:         800100,
	Name:          "sqlgen.dbcol",
	Tag:           "bytes,800100,opt,name=dbcol",
	Filename:      "sqlgen/eyevip.proto",
}

var E_Dbpk = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.FieldOptions)(nil),
	ExtensionType: (*string)(nil),
	Field:         800101,
	Name:          "sqlgen.dbpk",
	Tag:           "bytes,800101,opt,name=dbpk",
	Filename:      "sqlgen/eyevip.proto",
}

func init() {
	proto.RegisterExtension(E_Dbcol)
	proto.RegisterExtension(E_Dbpk)
}

func init() { proto.RegisterFile("sqlgen/eyevip.proto", fileDescriptor_c68ba2f54c05f7f7) }

var fileDescriptor_c68ba2f54c05f7f7 = []byte{
	// 151 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2e, 0x2e, 0xcc, 0x49,
	0x4f, 0xcd, 0xd3, 0x4f, 0xad, 0x4c, 0x2d, 0xcb, 0x2c, 0xd0, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17,
	0x62, 0x83, 0x08, 0x4a, 0x29, 0xa4, 0xe7, 0xe7, 0xa7, 0xe7, 0xa4, 0xea, 0x83, 0x45, 0x93, 0x4a,
	0xd3, 0xf4, 0x53, 0x52, 0x8b, 0x93, 0x8b, 0x32, 0x0b, 0x4a, 0xf2, 0x8b, 0x20, 0x2a, 0xad, 0x4c,
	0xb9, 0x58, 0x53, 0x92, 0x92, 0xf3, 0x73, 0x84, 0x64, 0xf5, 0x20, 0x6a, 0xf5, 0x60, 0x6a, 0xf5,
	0xdc, 0x32, 0x53, 0x73, 0x52, 0xfc, 0x0b, 0x4a, 0x32, 0xf3, 0xf3, 0x8a, 0x25, 0x9e, 0xbc, 0x32,
	0x50, 0x60, 0xd4, 0xe0, 0x0c, 0x82, 0xa8, 0xb6, 0x32, 0xe6, 0x62, 0x49, 0x49, 0x2a, 0xc8, 0x26,
	0xa4, 0xeb, 0x29, 0x54, 0x17, 0x58, 0xb1, 0x93, 0xc0, 0x89, 0x47, 0x72, 0x8c, 0x17, 0x1e, 0xc9,
	0x31, 0x3e, 0x78, 0x24, 0xc7, 0x38, 0xe3, 0xb1, 0x1c, 0x03, 0x20, 0x00, 0x00, 0xff, 0xff, 0x9f,
	0x3a, 0xb9, 0x6a, 0xbd, 0x00, 0x00, 0x00,
}
