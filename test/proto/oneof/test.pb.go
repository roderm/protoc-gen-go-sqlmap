// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: test/proto/oneof/test.proto

package oneof

import (
	_ "github.com/roderm/protoc-gen-go-sqlmap/sqlgen"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Product struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProductID   string `protobuf:"bytes,1,opt,name=ProductID,proto3" json:"ProductID,omitempty" db:"product_id"`
	ProductName string `protobuf:"bytes,2,opt,name=ProductName,proto3" json:"ProductName,omitempty" db:"product_name"`
	ProductType int32  `protobuf:"varint,3,opt,name=ProductType,proto3" json:"ProductType,omitempty" db:"product_type"`
	// Types that are assignable to Type:
	//	*Product_Software
	//	*Product_Hardware
	//	*Product_Service
	Type isProduct_Type `protobuf_oneof:"Type"`
}

func (x *Product) Reset() {
	*x = Product{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_proto_oneof_test_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Product) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Product) ProtoMessage() {}

func (x *Product) ProtoReflect() protoreflect.Message {
	mi := &file_test_proto_oneof_test_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Product.ProtoReflect.Descriptor instead.
func (*Product) Descriptor() ([]byte, []int) {
	return file_test_proto_oneof_test_proto_rawDescGZIP(), []int{0}
}

func (x *Product) GetProductID() string {
	if x != nil {
		return x.ProductID
	}
	return ""
}

func (x *Product) GetProductName() string {
	if x != nil {
		return x.ProductName
	}
	return ""
}

func (x *Product) GetProductType() int32 {
	if x != nil {
		return x.ProductType
	}
	return 0
}

func (m *Product) GetType() isProduct_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Product) GetSoftware() *Software {
	if x, ok := x.GetType().(*Product_Software); ok {
		return x.Software
	}
	return nil
}

func (x *Product) GetHardware() *Hardware {
	if x, ok := x.GetType().(*Product_Hardware); ok {
		return x.Hardware
	}
	return nil
}

func (x *Product) GetService() *Service {
	if x, ok := x.GetType().(*Product_Service); ok {
		return x.Service
	}
	return nil
}

type isProduct_Type interface {
	isProduct_Type()
}

type Product_Software struct {
	Software *Software `protobuf:"bytes,4,opt,name=software,proto3,oneof"`
}

type Product_Hardware struct {
	Hardware *Hardware `protobuf:"bytes,5,opt,name=hardware,proto3,oneof"`
}

type Product_Service struct {
	Service *Service `protobuf:"bytes,6,opt,name=service,proto3,oneof"`
}

func (*Product_Software) isProduct_Type() {}

func (*Product_Hardware) isProduct_Type() {}

func (*Product_Service) isProduct_Type() {}

type Software struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProductID string `protobuf:"bytes,1,opt,name=ProductID,proto3" json:"ProductID,omitempty" db:"product_id"`
	Version   string `protobuf:"bytes,2,opt,name=Version,proto3" json:"Version,omitempty" db:"product_version"`
}

func (x *Software) Reset() {
	*x = Software{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_proto_oneof_test_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Software) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Software) ProtoMessage() {}

func (x *Software) ProtoReflect() protoreflect.Message {
	mi := &file_test_proto_oneof_test_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Software.ProtoReflect.Descriptor instead.
func (*Software) Descriptor() ([]byte, []int) {
	return file_test_proto_oneof_test_proto_rawDescGZIP(), []int{1}
}

func (x *Software) GetProductID() string {
	if x != nil {
		return x.ProductID
	}
	return ""
}

func (x *Software) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

type Hardware struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProductID string `protobuf:"bytes,1,opt,name=ProductID,proto3" json:"ProductID,omitempty" db:"product_id"`
	Serial    string `protobuf:"bytes,2,opt,name=Serial,proto3" json:"Serial,omitempty" db:"product_serial"`
}

func (x *Hardware) Reset() {
	*x = Hardware{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_proto_oneof_test_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Hardware) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Hardware) ProtoMessage() {}

func (x *Hardware) ProtoReflect() protoreflect.Message {
	mi := &file_test_proto_oneof_test_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Hardware.ProtoReflect.Descriptor instead.
func (*Hardware) Descriptor() ([]byte, []int) {
	return file_test_proto_oneof_test_proto_rawDescGZIP(), []int{2}
}

func (x *Hardware) GetProductID() string {
	if x != nil {
		return x.ProductID
	}
	return ""
}

func (x *Hardware) GetSerial() string {
	if x != nil {
		return x.Serial
	}
	return ""
}

type Service struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProductID string `protobuf:"bytes,1,opt,name=ProductID,proto3" json:"ProductID,omitempty" db:"product_id"`
}

func (x *Service) Reset() {
	*x = Service{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_proto_oneof_test_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Service) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Service) ProtoMessage() {}

func (x *Service) ProtoReflect() protoreflect.Message {
	mi := &file_test_proto_oneof_test_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Service.ProtoReflect.Descriptor instead.
func (*Service) Descriptor() ([]byte, []int) {
	return file_test_proto_oneof_test_proto_rawDescGZIP(), []int{3}
}

func (x *Service) GetProductID() string {
	if x != nil {
		return x.ProductID
	}
	return ""
}

var File_test_proto_oneof_test_proto protoreflect.FileDescriptor

var file_test_proto_oneof_test_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x6e, 0x65,
	0x6f, 0x66, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x6f,
	0x6e, 0x65, 0x6f, 0x66, 0x1a, 0x13, 0x73, 0x71, 0x6c, 0x67, 0x65, 0x6e, 0x2f, 0x73, 0x71, 0x6c,
	0x67, 0x65, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd5, 0x03, 0x0a, 0x07, 0x50, 0x72,
	0x6f, 0x64, 0x75, 0x63, 0x74, 0x12, 0x31, 0x0a, 0x09, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74,
	0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x13, 0xf2, 0xd6, 0x86, 0x03, 0x0e, 0x0a,
	0x0a, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x10, 0x01, 0x52, 0x09, 0x50,
	0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x49, 0x44, 0x12, 0x35, 0x0a, 0x0b, 0x50, 0x72, 0x6f, 0x64,
	0x75, 0x63, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x13, 0xf2,
	0xd6, 0x86, 0x03, 0x0e, 0x0a, 0x0c, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x52, 0x0b, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12,
	0x35, 0x0a, 0x0b, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x54, 0x79, 0x70, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x05, 0x42, 0x13, 0xf2, 0xd6, 0x86, 0x03, 0x0e, 0x0a, 0x0c, 0x70, 0x72, 0x6f,
	0x64, 0x75, 0x63, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x52, 0x0b, 0x50, 0x72, 0x6f, 0x64, 0x75,
	0x63, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x59, 0x0a, 0x08, 0x73, 0x6f, 0x66, 0x74, 0x77, 0x61,
	0x72, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x6f, 0x6e, 0x65, 0x6f, 0x66,
	0x2e, 0x53, 0x6f, 0x66, 0x74, 0x77, 0x61, 0x72, 0x65, 0x42, 0x2a, 0xf2, 0xd6, 0x86, 0x03, 0x25,
	0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x1a, 0x17, 0x74, 0x62,
	0x6c, 0x5f, 0x73, 0x6f, 0x66, 0x74, 0x77, 0x61, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75,
	0x63, 0x74, 0x5f, 0x69, 0x64, 0x48, 0x00, 0x52, 0x08, 0x73, 0x6f, 0x66, 0x74, 0x77, 0x61, 0x72,
	0x65, 0x12, 0x59, 0x0a, 0x08, 0x68, 0x61, 0x72, 0x64, 0x77, 0x61, 0x72, 0x65, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x2e, 0x48, 0x61, 0x72, 0x64,
	0x77, 0x61, 0x72, 0x65, 0x42, 0x2a, 0xf2, 0xd6, 0x86, 0x03, 0x25, 0x0a, 0x0a, 0x70, 0x72, 0x6f,
	0x64, 0x75, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x1a, 0x17, 0x74, 0x62, 0x6c, 0x5f, 0x68, 0x61, 0x72,
	0x64, 0x77, 0x61, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f, 0x69, 0x64,
	0x48, 0x00, 0x52, 0x08, 0x68, 0x61, 0x72, 0x64, 0x77, 0x61, 0x72, 0x65, 0x12, 0x55, 0x0a, 0x07,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e,
	0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x42, 0x29, 0xf2,
	0xd6, 0x86, 0x03, 0x24, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f, 0x69, 0x64,
	0x1a, 0x16, 0x74, 0x62, 0x6c, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x48, 0x00, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x3a, 0x14, 0xa2, 0xd6, 0x86, 0x03, 0x0f, 0x0a, 0x0b, 0x74, 0x62, 0x6c, 0x5f,
	0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x10, 0x01, 0x42, 0x06, 0x0a, 0x04, 0x54, 0x79, 0x70,
	0x65, 0x22, 0x86, 0x01, 0x0a, 0x08, 0x53, 0x6f, 0x66, 0x74, 0x77, 0x61, 0x72, 0x65, 0x12, 0x31,
	0x0a, 0x09, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x13, 0xf2, 0xd6, 0x86, 0x03, 0x0e, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63,
	0x74, 0x5f, 0x69, 0x64, 0x10, 0x02, 0x52, 0x09, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x49,
	0x44, 0x12, 0x30, 0x0a, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x16, 0xf2, 0xd6, 0x86, 0x03, 0x11, 0x0a, 0x0f, 0x70, 0x72, 0x6f, 0x64, 0x75,
	0x63, 0x74, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x07, 0x56, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x3a, 0x15, 0xa2, 0xd6, 0x86, 0x03, 0x10, 0x0a, 0x0c, 0x74, 0x62, 0x6c, 0x5f,
	0x73, 0x6f, 0x66, 0x74, 0x77, 0x61, 0x72, 0x65, 0x10, 0x01, 0x22, 0x83, 0x01, 0x0a, 0x08, 0x48,
	0x61, 0x72, 0x64, 0x77, 0x61, 0x72, 0x65, 0x12, 0x31, 0x0a, 0x09, 0x50, 0x72, 0x6f, 0x64, 0x75,
	0x63, 0x74, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x13, 0xf2, 0xd6, 0x86, 0x03,
	0x0e, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x10, 0x02, 0x52,
	0x09, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x49, 0x44, 0x12, 0x2d, 0x0a, 0x06, 0x53, 0x65,
	0x72, 0x69, 0x61, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x15, 0xf2, 0xd6, 0x86, 0x03,
	0x10, 0x0a, 0x0e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f, 0x73, 0x65, 0x72, 0x69, 0x61,
	0x6c, 0x52, 0x06, 0x53, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x3a, 0x15, 0xa2, 0xd6, 0x86, 0x03, 0x10,
	0x0a, 0x0c, 0x74, 0x62, 0x6c, 0x5f, 0x68, 0x61, 0x72, 0x64, 0x77, 0x61, 0x72, 0x65, 0x10, 0x01,
	0x22, 0x52, 0x0a, 0x07, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x31, 0x0a, 0x09, 0x50,
	0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x13,
	0xf2, 0xd6, 0x86, 0x03, 0x0e, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f, 0x69,
	0x64, 0x10, 0x02, 0x52, 0x09, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x49, 0x44, 0x3a, 0x14,
	0xa2, 0xd6, 0x86, 0x03, 0x0f, 0x0a, 0x0b, 0x74, 0x62, 0x6c, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x10, 0x01, 0x42, 0x39, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x72, 0x6f, 0x64, 0x65, 0x72, 0x6d, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x67, 0x6f, 0x2d, 0x73, 0x71, 0x6c, 0x6d, 0x61, 0x70, 0x2f, 0x74,
	0x65, 0x73, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_test_proto_oneof_test_proto_rawDescOnce sync.Once
	file_test_proto_oneof_test_proto_rawDescData = file_test_proto_oneof_test_proto_rawDesc
)

func file_test_proto_oneof_test_proto_rawDescGZIP() []byte {
	file_test_proto_oneof_test_proto_rawDescOnce.Do(func() {
		file_test_proto_oneof_test_proto_rawDescData = protoimpl.X.CompressGZIP(file_test_proto_oneof_test_proto_rawDescData)
	})
	return file_test_proto_oneof_test_proto_rawDescData
}

var file_test_proto_oneof_test_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_test_proto_oneof_test_proto_goTypes = []interface{}{
	(*Product)(nil),  // 0: oneof.Product
	(*Software)(nil), // 1: oneof.Software
	(*Hardware)(nil), // 2: oneof.Hardware
	(*Service)(nil),  // 3: oneof.Service
}
var file_test_proto_oneof_test_proto_depIdxs = []int32{
	1, // 0: oneof.Product.software:type_name -> oneof.Software
	2, // 1: oneof.Product.hardware:type_name -> oneof.Hardware
	3, // 2: oneof.Product.service:type_name -> oneof.Service
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_test_proto_oneof_test_proto_init() }
func file_test_proto_oneof_test_proto_init() {
	if File_test_proto_oneof_test_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_test_proto_oneof_test_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Product); i {
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
		file_test_proto_oneof_test_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Software); i {
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
		file_test_proto_oneof_test_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Hardware); i {
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
		file_test_proto_oneof_test_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Service); i {
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
	file_test_proto_oneof_test_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Product_Software)(nil),
		(*Product_Hardware)(nil),
		(*Product_Service)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_test_proto_oneof_test_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_test_proto_oneof_test_proto_goTypes,
		DependencyIndexes: file_test_proto_oneof_test_proto_depIdxs,
		MessageInfos:      file_test_proto_oneof_test_proto_msgTypes,
	}.Build()
	File_test_proto_oneof_test_proto = out.File
	file_test_proto_oneof_test_proto_rawDesc = nil
	file_test_proto_oneof_test_proto_goTypes = nil
	file_test_proto_oneof_test_proto_depIdxs = nil
}
