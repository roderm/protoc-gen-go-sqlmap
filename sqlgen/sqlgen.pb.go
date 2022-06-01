// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        (unknown)
// source: sqlgen/sqlgen.proto

package sqlgen

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PK int32

const (
	PK_PK_UNSPECIFIED PK = 0
	PK_PK_AUTO        PK = 1
	PK_PK_MAN         PK = 2
)

// Enum value maps for PK.
var (
	PK_name = map[int32]string{
		0: "PK_UNSPECIFIED",
		1: "PK_AUTO",
		2: "PK_MAN",
	}
	PK_value = map[string]int32{
		"PK_UNSPECIFIED": 0,
		"PK_AUTO":        1,
		"PK_MAN":         2,
	}
)

func (x PK) Enum() *PK {
	p := new(PK)
	*p = x
	return p
}

func (x PK) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PK) Descriptor() protoreflect.EnumDescriptor {
	return file_sqlgen_sqlgen_proto_enumTypes[0].Descriptor()
}

func (PK) Type() protoreflect.EnumType {
	return &file_sqlgen_sqlgen_proto_enumTypes[0]
}

func (x PK) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *PK) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = PK(num)
	return nil
}

// Deprecated: Use PK.Descriptor instead.
func (PK) EnumDescriptor() ([]byte, []int) {
	return file_sqlgen_sqlgen_proto_rawDescGZIP(), []int{0}
}

type OPERATION int32

const (
	OPERATION_C OPERATION = 0
	OPERATION_R OPERATION = 1
	OPERATION_U OPERATION = 2
	OPERATION_D OPERATION = 3
)

// Enum value maps for OPERATION.
var (
	OPERATION_name = map[int32]string{
		0: "C",
		1: "R",
		2: "U",
		3: "D",
	}
	OPERATION_value = map[string]int32{
		"C": 0,
		"R": 1,
		"U": 2,
		"D": 3,
	}
)

func (x OPERATION) Enum() *OPERATION {
	p := new(OPERATION)
	*p = x
	return p
}

func (x OPERATION) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (OPERATION) Descriptor() protoreflect.EnumDescriptor {
	return file_sqlgen_sqlgen_proto_enumTypes[1].Descriptor()
}

func (OPERATION) Type() protoreflect.EnumType {
	return &file_sqlgen_sqlgen_proto_enumTypes[1]
}

func (x OPERATION) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *OPERATION) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = OPERATION(num)
	return nil
}

// Deprecated: Use OPERATION.Descriptor instead.
func (OPERATION) EnumDescriptor() ([]byte, []int) {
	return file_sqlgen_sqlgen_proto_rawDescGZIP(), []int{1}
}

type Table struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name *string     `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Crud []OPERATION `protobuf:"varint,2,rep,name=crud,enum=sqlgen.OPERATION" json:"crud,omitempty"`
}

func (x *Table) Reset() {
	*x = Table{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sqlgen_sqlgen_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Table) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Table) ProtoMessage() {}

func (x *Table) ProtoReflect() protoreflect.Message {
	mi := &file_sqlgen_sqlgen_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Table.ProtoReflect.Descriptor instead.
func (*Table) Descriptor() ([]byte, []int) {
	return file_sqlgen_sqlgen_proto_rawDescGZIP(), []int{0}
}

func (x *Table) GetName() string {
	if x != nil && x.Name != nil {
		return *x.Name
	}
	return ""
}

func (x *Table) GetCrud() []OPERATION {
	if x != nil {
		return x.Crud
	}
	return nil
}

type ForeignCol struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Src *string `protobuf:"bytes,1,req,name=src" json:"src,omitempty"`
	Dst *string `protobuf:"bytes,2,req,name=dst" json:"dst,omitempty"`
}

func (x *ForeignCol) Reset() {
	*x = ForeignCol{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sqlgen_sqlgen_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ForeignCol) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ForeignCol) ProtoMessage() {}

func (x *ForeignCol) ProtoReflect() protoreflect.Message {
	mi := &file_sqlgen_sqlgen_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ForeignCol.ProtoReflect.Descriptor instead.
func (*ForeignCol) Descriptor() ([]byte, []int) {
	return file_sqlgen_sqlgen_proto_rawDescGZIP(), []int{1}
}

func (x *ForeignCol) GetSrc() string {
	if x != nil && x.Src != nil {
		return *x.Src
	}
	return ""
}

func (x *ForeignCol) GetDst() string {
	if x != nil && x.Dst != nil {
		return *x.Dst
	}
	return ""
}

type ForeignKey struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Table *string       `protobuf:"bytes,1,req,name=table" json:"table,omitempty"`
	Field []*ForeignCol `protobuf:"bytes,2,rep,name=field" json:"field,omitempty"`
}

func (x *ForeignKey) Reset() {
	*x = ForeignKey{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sqlgen_sqlgen_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ForeignKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ForeignKey) ProtoMessage() {}

func (x *ForeignKey) ProtoReflect() protoreflect.Message {
	mi := &file_sqlgen_sqlgen_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ForeignKey.ProtoReflect.Descriptor instead.
func (*ForeignKey) Descriptor() ([]byte, []int) {
	return file_sqlgen_sqlgen_proto_rawDescGZIP(), []int{2}
}

func (x *ForeignKey) GetTable() string {
	if x != nil && x.Table != nil {
		return *x.Table
	}
	return ""
}

func (x *ForeignKey) GetField() []*ForeignCol {
	if x != nil {
		return x.Field
	}
	return nil
}

type Column struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name *string `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Pk   *PK     `protobuf:"varint,2,opt,name=pk,enum=sqlgen.PK" json:"pk,omitempty"`
	Fk   *string `protobuf:"bytes,3,opt,name=fk" json:"fk,omitempty"`
}

func (x *Column) Reset() {
	*x = Column{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sqlgen_sqlgen_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Column) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Column) ProtoMessage() {}

func (x *Column) ProtoReflect() protoreflect.Message {
	mi := &file_sqlgen_sqlgen_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Column.ProtoReflect.Descriptor instead.
func (*Column) Descriptor() ([]byte, []int) {
	return file_sqlgen_sqlgen_proto_rawDescGZIP(), []int{3}
}

func (x *Column) GetName() string {
	if x != nil && x.Name != nil {
		return *x.Name
	}
	return ""
}

func (x *Column) GetPk() PK {
	if x != nil && x.Pk != nil {
		return *x.Pk
	}
	return PK_PK_UNSPECIFIED
}

func (x *Column) GetFk() string {
	if x != nil && x.Fk != nil {
		return *x.Fk
	}
	return ""
}

var file_sqlgen_sqlgen_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.FileOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         800200,
		Name:          "sqlgen.sqlgen",
		Tag:           "varint,800200,opt,name=sqlgen",
		Filename:      "sqlgen/sqlgen.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FileOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         800201,
		Name:          "sqlgen.StoreName",
		Tag:           "bytes,800201,opt,name=StoreName",
		Filename:      "sqlgen/sqlgen.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MessageOptions)(nil),
		ExtensionType: (*Table)(nil),
		Field:         800100,
		Name:          "sqlgen.table",
		Tag:           "bytes,800100,opt,name=table",
		Filename:      "sqlgen/sqlgen.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MessageOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         800101,
		Name:          "sqlgen.jsonb",
		Tag:           "varint,800101,opt,name=jsonb",
		Filename:      "sqlgen/sqlgen.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*Column)(nil),
		Field:         800110,
		Name:          "sqlgen.col",
		Tag:           "bytes,800110,opt,name=col",
		Filename:      "sqlgen/sqlgen.proto",
	},
}

// Extension fields to descriptorpb.FileOptions.
var (
	// optional bool sqlgen = 800200;
	E_Sqlgen = &file_sqlgen_sqlgen_proto_extTypes[0]
	// optional string StoreName = 800201;
	E_StoreName = &file_sqlgen_sqlgen_proto_extTypes[1]
)

// Extension fields to descriptorpb.MessageOptions.
var (
	// optional sqlgen.Table table = 800100;
	E_Table = &file_sqlgen_sqlgen_proto_extTypes[2]
	// optional bool jsonb = 800101;
	E_Jsonb = &file_sqlgen_sqlgen_proto_extTypes[3] // optional CRUD crud = 800101;
)

// Extension fields to descriptorpb.FieldOptions.
var (
	// optional sqlgen.Column col = 800110;
	E_Col = &file_sqlgen_sqlgen_proto_extTypes[4]
)

var File_sqlgen_sqlgen_proto protoreflect.FileDescriptor

var file_sqlgen_sqlgen_proto_rawDesc = []byte{
	0x0a, 0x13, 0x73, 0x71, 0x6c, 0x67, 0x65, 0x6e, 0x2f, 0x73, 0x71, 0x6c, 0x67, 0x65, 0x6e, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x73, 0x71, 0x6c, 0x67, 0x65, 0x6e, 0x1a, 0x20, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64,
	0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x42, 0x0a, 0x05, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x02, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x25, 0x0a, 0x04,
	0x63, 0x72, 0x75, 0x64, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x11, 0x2e, 0x73, 0x71, 0x6c,
	0x67, 0x65, 0x6e, 0x2e, 0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x52, 0x04, 0x63,
	0x72, 0x75, 0x64, 0x22, 0x30, 0x0a, 0x0a, 0x46, 0x6f, 0x72, 0x65, 0x69, 0x67, 0x6e, 0x43, 0x6f,
	0x6c, 0x12, 0x10, 0x0a, 0x03, 0x73, 0x72, 0x63, 0x18, 0x01, 0x20, 0x02, 0x28, 0x09, 0x52, 0x03,
	0x73, 0x72, 0x63, 0x12, 0x10, 0x0a, 0x03, 0x64, 0x73, 0x74, 0x18, 0x02, 0x20, 0x02, 0x28, 0x09,
	0x52, 0x03, 0x64, 0x73, 0x74, 0x22, 0x4c, 0x0a, 0x0a, 0x46, 0x6f, 0x72, 0x65, 0x69, 0x67, 0x6e,
	0x4b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x02,
	0x28, 0x09, 0x52, 0x05, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x66, 0x69, 0x65,
	0x6c, 0x64, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x73, 0x71, 0x6c, 0x67, 0x65,
	0x6e, 0x2e, 0x46, 0x6f, 0x72, 0x65, 0x69, 0x67, 0x6e, 0x43, 0x6f, 0x6c, 0x52, 0x05, 0x66, 0x69,
	0x65, 0x6c, 0x64, 0x22, 0x48, 0x0a, 0x06, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x02, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x1a, 0x0a, 0x02, 0x70, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0a, 0x2e,
	0x73, 0x71, 0x6c, 0x67, 0x65, 0x6e, 0x2e, 0x50, 0x4b, 0x52, 0x02, 0x70, 0x6b, 0x12, 0x0e, 0x0a,
	0x02, 0x66, 0x6b, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x66, 0x6b, 0x2a, 0x31, 0x0a,
	0x02, 0x50, 0x4b, 0x12, 0x12, 0x0a, 0x0e, 0x50, 0x4b, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43,
	0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x50, 0x4b, 0x5f, 0x41, 0x55,
	0x54, 0x4f, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x50, 0x4b, 0x5f, 0x4d, 0x41, 0x4e, 0x10, 0x02,
	0x2a, 0x27, 0x0a, 0x09, 0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x12, 0x05, 0x0a,
	0x01, 0x43, 0x10, 0x00, 0x12, 0x05, 0x0a, 0x01, 0x52, 0x10, 0x01, 0x12, 0x05, 0x0a, 0x01, 0x55,
	0x10, 0x02, 0x12, 0x05, 0x0a, 0x01, 0x44, 0x10, 0x03, 0x3a, 0x36, 0x0a, 0x06, 0x73, 0x71, 0x6c,
	0x67, 0x65, 0x6e, 0x12, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x18, 0xc8, 0xeb, 0x30, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x73, 0x71, 0x6c, 0x67, 0x65,
	0x6e, 0x3a, 0x3c, 0x0a, 0x09, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1c,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x46, 0x69, 0x6c, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xc9, 0xeb, 0x30,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x3a,
	0x46, 0x0a, 0x05, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xe4, 0xea, 0x30, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0d, 0x2e, 0x73, 0x71, 0x6c, 0x67, 0x65, 0x6e, 0x2e, 0x54, 0x61, 0x62, 0x6c, 0x65,
	0x52, 0x05, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x3a, 0x37, 0x0a, 0x05, 0x6a, 0x73, 0x6f, 0x6e, 0x62,
	0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x18, 0xe5, 0xea, 0x30, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x6a, 0x73, 0x6f, 0x6e, 0x62,
	0x3a, 0x41, 0x0a, 0x03, 0x63, 0x6f, 0x6c, 0x12, 0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xee, 0xea, 0x30, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e,
	0x2e, 0x73, 0x71, 0x6c, 0x67, 0x65, 0x6e, 0x2e, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x52, 0x03,
	0x63, 0x6f, 0x6c, 0x42, 0x80, 0x01, 0x0a, 0x0a, 0x63, 0x6f, 0x6d, 0x2e, 0x73, 0x71, 0x6c, 0x67,
	0x65, 0x6e, 0x42, 0x0b, 0x53, 0x71, 0x6c, 0x67, 0x65, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50,
	0x01, 0x5a, 0x2d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x72, 0x6f,
	0x64, 0x65, 0x72, 0x6d, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d,
	0x67, 0x6f, 0x2d, 0x73, 0x71, 0x6c, 0x6d, 0x61, 0x70, 0x2f, 0x73, 0x71, 0x6c, 0x67, 0x65, 0x6e,
	0xa2, 0x02, 0x03, 0x53, 0x58, 0x58, 0xaa, 0x02, 0x06, 0x53, 0x71, 0x6c, 0x67, 0x65, 0x6e, 0xca,
	0x02, 0x06, 0x53, 0x71, 0x6c, 0x67, 0x65, 0x6e, 0xe2, 0x02, 0x12, 0x53, 0x71, 0x6c, 0x67, 0x65,
	0x6e, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x06,
	0x53, 0x71, 0x6c, 0x67, 0x65, 0x6e,
}

var (
	file_sqlgen_sqlgen_proto_rawDescOnce sync.Once
	file_sqlgen_sqlgen_proto_rawDescData = file_sqlgen_sqlgen_proto_rawDesc
)

func file_sqlgen_sqlgen_proto_rawDescGZIP() []byte {
	file_sqlgen_sqlgen_proto_rawDescOnce.Do(func() {
		file_sqlgen_sqlgen_proto_rawDescData = protoimpl.X.CompressGZIP(file_sqlgen_sqlgen_proto_rawDescData)
	})
	return file_sqlgen_sqlgen_proto_rawDescData
}

var file_sqlgen_sqlgen_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_sqlgen_sqlgen_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_sqlgen_sqlgen_proto_goTypes = []interface{}{
	(PK)(0),                             // 0: sqlgen.PK
	(OPERATION)(0),                      // 1: sqlgen.OPERATION
	(*Table)(nil),                       // 2: sqlgen.Table
	(*ForeignCol)(nil),                  // 3: sqlgen.ForeignCol
	(*ForeignKey)(nil),                  // 4: sqlgen.ForeignKey
	(*Column)(nil),                      // 5: sqlgen.Column
	(*descriptorpb.FileOptions)(nil),    // 6: google.protobuf.FileOptions
	(*descriptorpb.MessageOptions)(nil), // 7: google.protobuf.MessageOptions
	(*descriptorpb.FieldOptions)(nil),   // 8: google.protobuf.FieldOptions
}
var file_sqlgen_sqlgen_proto_depIdxs = []int32{
	1,  // 0: sqlgen.Table.crud:type_name -> sqlgen.OPERATION
	3,  // 1: sqlgen.ForeignKey.field:type_name -> sqlgen.ForeignCol
	0,  // 2: sqlgen.Column.pk:type_name -> sqlgen.PK
	6,  // 3: sqlgen.sqlgen:extendee -> google.protobuf.FileOptions
	6,  // 4: sqlgen.StoreName:extendee -> google.protobuf.FileOptions
	7,  // 5: sqlgen.table:extendee -> google.protobuf.MessageOptions
	7,  // 6: sqlgen.jsonb:extendee -> google.protobuf.MessageOptions
	8,  // 7: sqlgen.col:extendee -> google.protobuf.FieldOptions
	2,  // 8: sqlgen.table:type_name -> sqlgen.Table
	5,  // 9: sqlgen.col:type_name -> sqlgen.Column
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	8,  // [8:10] is the sub-list for extension type_name
	3,  // [3:8] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_sqlgen_sqlgen_proto_init() }
func file_sqlgen_sqlgen_proto_init() {
	if File_sqlgen_sqlgen_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_sqlgen_sqlgen_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Table); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_sqlgen_sqlgen_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ForeignCol); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_sqlgen_sqlgen_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ForeignKey); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_sqlgen_sqlgen_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Column); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_sqlgen_sqlgen_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   4,
			NumExtensions: 5,
			NumServices:   0,
		},
		GoTypes:           file_sqlgen_sqlgen_proto_goTypes,
		DependencyIndexes: file_sqlgen_sqlgen_proto_depIdxs,
		EnumInfos:         file_sqlgen_sqlgen_proto_enumTypes,
		MessageInfos:      file_sqlgen_sqlgen_proto_msgTypes,
		ExtensionInfos:    file_sqlgen_sqlgen_proto_extTypes,
	}.Build()
	File_sqlgen_sqlgen_proto = out.File
	file_sqlgen_sqlgen_proto_rawDesc = nil
	file_sqlgen_sqlgen_proto_goTypes = nil
	file_sqlgen_sqlgen_proto_depIdxs = nil
}
