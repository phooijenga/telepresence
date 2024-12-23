// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.0
// 	protoc        v3.21.9
// source: common/errors.proto

package common

import (
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

// InterceptError is a common error type used by the intercept call family (add,
// remove, list, available).
type InterceptError int32

const (
	InterceptError_UNSPECIFIED                InterceptError = 0 // no error
	InterceptError_INTERNAL                   InterceptError = 1
	InterceptError_NO_CONNECTION              InterceptError = 2 // Have not made the .Connect RPC call (or it errored)
	InterceptError_NO_TRAFFIC_MANAGER         InterceptError = 3
	InterceptError_TRAFFIC_MANAGER_CONNECTING InterceptError = 4
	InterceptError_TRAFFIC_MANAGER_ERROR      InterceptError = 5
	InterceptError_ALREADY_EXISTS             InterceptError = 6
	InterceptError_NAMESPACE_AMBIGUITY        InterceptError = 17
	InterceptError_LOCAL_TARGET_IN_USE        InterceptError = 7
	InterceptError_NO_ACCEPTABLE_WORKLOAD     InterceptError = 8
	InterceptError_AMBIGUOUS_MATCH            InterceptError = 9
	InterceptError_FAILED_TO_ESTABLISH        InterceptError = 10
	InterceptError_UNSUPPORTED_WORKLOAD       InterceptError = 11
	InterceptError_MISCONFIGURED_WORKLOAD     InterceptError = 14
	InterceptError_NOT_FOUND                  InterceptError = 12
	InterceptError_MOUNT_POINT_BUSY           InterceptError = 13
	InterceptError_UNKNOWN_FLAG               InterceptError = 15
	InterceptError_EXEC_CMD                   InterceptError = 16 // External exec command failed
)

// Enum value maps for InterceptError.
var (
	InterceptError_name = map[int32]string{
		0:  "UNSPECIFIED",
		1:  "INTERNAL",
		2:  "NO_CONNECTION",
		3:  "NO_TRAFFIC_MANAGER",
		4:  "TRAFFIC_MANAGER_CONNECTING",
		5:  "TRAFFIC_MANAGER_ERROR",
		6:  "ALREADY_EXISTS",
		17: "NAMESPACE_AMBIGUITY",
		7:  "LOCAL_TARGET_IN_USE",
		8:  "NO_ACCEPTABLE_WORKLOAD",
		9:  "AMBIGUOUS_MATCH",
		10: "FAILED_TO_ESTABLISH",
		11: "UNSUPPORTED_WORKLOAD",
		14: "MISCONFIGURED_WORKLOAD",
		12: "NOT_FOUND",
		13: "MOUNT_POINT_BUSY",
		15: "UNKNOWN_FLAG",
		16: "EXEC_CMD",
	}
	InterceptError_value = map[string]int32{
		"UNSPECIFIED":                0,
		"INTERNAL":                   1,
		"NO_CONNECTION":              2,
		"NO_TRAFFIC_MANAGER":         3,
		"TRAFFIC_MANAGER_CONNECTING": 4,
		"TRAFFIC_MANAGER_ERROR":      5,
		"ALREADY_EXISTS":             6,
		"NAMESPACE_AMBIGUITY":        17,
		"LOCAL_TARGET_IN_USE":        7,
		"NO_ACCEPTABLE_WORKLOAD":     8,
		"AMBIGUOUS_MATCH":            9,
		"FAILED_TO_ESTABLISH":        10,
		"UNSUPPORTED_WORKLOAD":       11,
		"MISCONFIGURED_WORKLOAD":     14,
		"NOT_FOUND":                  12,
		"MOUNT_POINT_BUSY":           13,
		"UNKNOWN_FLAG":               15,
		"EXEC_CMD":                   16,
	}
)

func (x InterceptError) Enum() *InterceptError {
	p := new(InterceptError)
	*p = x
	return p
}

func (x InterceptError) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (InterceptError) Descriptor() protoreflect.EnumDescriptor {
	return file_common_errors_proto_enumTypes[0].Descriptor()
}

func (InterceptError) Type() protoreflect.EnumType {
	return &file_common_errors_proto_enumTypes[0]
}

func (x InterceptError) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use InterceptError.Descriptor instead.
func (InterceptError) EnumDescriptor() ([]byte, []int) {
	return file_common_errors_proto_rawDescGZIP(), []int{0}
}

type Result_ErrorCategory int32

const (
	Result_UNSPECIFIED    Result_ErrorCategory = 0 // No error
	Result_USER           Result_ErrorCategory = 1
	Result_CONFIG         Result_ErrorCategory = 2
	Result_NO_DAEMON_LOGS Result_ErrorCategory = 3
	Result_UNKNOWN        Result_ErrorCategory = 4
)

// Enum value maps for Result_ErrorCategory.
var (
	Result_ErrorCategory_name = map[int32]string{
		0: "UNSPECIFIED",
		1: "USER",
		2: "CONFIG",
		3: "NO_DAEMON_LOGS",
		4: "UNKNOWN",
	}
	Result_ErrorCategory_value = map[string]int32{
		"UNSPECIFIED":    0,
		"USER":           1,
		"CONFIG":         2,
		"NO_DAEMON_LOGS": 3,
		"UNKNOWN":        4,
	}
)

func (x Result_ErrorCategory) Enum() *Result_ErrorCategory {
	p := new(Result_ErrorCategory)
	*p = x
	return p
}

func (x Result_ErrorCategory) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Result_ErrorCategory) Descriptor() protoreflect.EnumDescriptor {
	return file_common_errors_proto_enumTypes[1].Descriptor()
}

func (Result_ErrorCategory) Type() protoreflect.EnumType {
	return &file_common_errors_proto_enumTypes[1]
}

func (x Result_ErrorCategory) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Result_ErrorCategory.Descriptor instead.
func (Result_ErrorCategory) EnumDescriptor() ([]byte, []int) {
	return file_common_errors_proto_rawDescGZIP(), []int{0, 0}
}

type Result struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Data          []byte                 `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	ErrorCategory Result_ErrorCategory   `protobuf:"varint,2,opt,name=error_category,json=errorCategory,proto3,enum=telepresence.common.Result_ErrorCategory" json:"error_category,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Result) Reset() {
	*x = Result{}
	mi := &file_common_errors_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Result) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Result) ProtoMessage() {}

func (x *Result) ProtoReflect() protoreflect.Message {
	mi := &file_common_errors_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Result.ProtoReflect.Descriptor instead.
func (*Result) Descriptor() ([]byte, []int) {
	return file_common_errors_proto_rawDescGZIP(), []int{0}
}

func (x *Result) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Result) GetErrorCategory() Result_ErrorCategory {
	if x != nil {
		return x.ErrorCategory
	}
	return Result_UNSPECIFIED
}

var File_common_errors_proto protoreflect.FileDescriptor

var file_common_errors_proto_rawDesc = []byte{
	0x0a, 0x13, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x72, 0x65, 0x73, 0x65,
	0x6e, 0x63, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x22, 0xc7, 0x01, 0x0a, 0x06, 0x52,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x50, 0x0a, 0x0e, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x5f, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x29, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x72, 0x65, 0x73, 0x65, 0x6e, 0x63, 0x65,
	0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x2e, 0x45,
	0x72, 0x72, 0x6f, 0x72, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x52, 0x0d, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x22, 0x57, 0x0a, 0x0d, 0x45,
	0x72, 0x72, 0x6f, 0x72, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x12, 0x0f, 0x0a, 0x0b,
	0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x08, 0x0a,
	0x04, 0x55, 0x53, 0x45, 0x52, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x43, 0x4f, 0x4e, 0x46, 0x49,
	0x47, 0x10, 0x02, 0x12, 0x12, 0x0a, 0x0e, 0x4e, 0x4f, 0x5f, 0x44, 0x41, 0x45, 0x4d, 0x4f, 0x4e,
	0x5f, 0x4c, 0x4f, 0x47, 0x53, 0x10, 0x03, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f,
	0x57, 0x4e, 0x10, 0x04, 0x2a, 0xa0, 0x03, 0x0a, 0x0e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x63, 0x65,
	0x70, 0x74, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x0f, 0x0a, 0x0b, 0x55, 0x4e, 0x53, 0x50, 0x45,
	0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x49, 0x4e, 0x54, 0x45,
	0x52, 0x4e, 0x41, 0x4c, 0x10, 0x01, 0x12, 0x11, 0x0a, 0x0d, 0x4e, 0x4f, 0x5f, 0x43, 0x4f, 0x4e,
	0x4e, 0x45, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x10, 0x02, 0x12, 0x16, 0x0a, 0x12, 0x4e, 0x4f, 0x5f,
	0x54, 0x52, 0x41, 0x46, 0x46, 0x49, 0x43, 0x5f, 0x4d, 0x41, 0x4e, 0x41, 0x47, 0x45, 0x52, 0x10,
	0x03, 0x12, 0x1e, 0x0a, 0x1a, 0x54, 0x52, 0x41, 0x46, 0x46, 0x49, 0x43, 0x5f, 0x4d, 0x41, 0x4e,
	0x41, 0x47, 0x45, 0x52, 0x5f, 0x43, 0x4f, 0x4e, 0x4e, 0x45, 0x43, 0x54, 0x49, 0x4e, 0x47, 0x10,
	0x04, 0x12, 0x19, 0x0a, 0x15, 0x54, 0x52, 0x41, 0x46, 0x46, 0x49, 0x43, 0x5f, 0x4d, 0x41, 0x4e,
	0x41, 0x47, 0x45, 0x52, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10, 0x05, 0x12, 0x12, 0x0a, 0x0e,
	0x41, 0x4c, 0x52, 0x45, 0x41, 0x44, 0x59, 0x5f, 0x45, 0x58, 0x49, 0x53, 0x54, 0x53, 0x10, 0x06,
	0x12, 0x17, 0x0a, 0x13, 0x4e, 0x41, 0x4d, 0x45, 0x53, 0x50, 0x41, 0x43, 0x45, 0x5f, 0x41, 0x4d,
	0x42, 0x49, 0x47, 0x55, 0x49, 0x54, 0x59, 0x10, 0x11, 0x12, 0x17, 0x0a, 0x13, 0x4c, 0x4f, 0x43,
	0x41, 0x4c, 0x5f, 0x54, 0x41, 0x52, 0x47, 0x45, 0x54, 0x5f, 0x49, 0x4e, 0x5f, 0x55, 0x53, 0x45,
	0x10, 0x07, 0x12, 0x1a, 0x0a, 0x16, 0x4e, 0x4f, 0x5f, 0x41, 0x43, 0x43, 0x45, 0x50, 0x54, 0x41,
	0x42, 0x4c, 0x45, 0x5f, 0x57, 0x4f, 0x52, 0x4b, 0x4c, 0x4f, 0x41, 0x44, 0x10, 0x08, 0x12, 0x13,
	0x0a, 0x0f, 0x41, 0x4d, 0x42, 0x49, 0x47, 0x55, 0x4f, 0x55, 0x53, 0x5f, 0x4d, 0x41, 0x54, 0x43,
	0x48, 0x10, 0x09, 0x12, 0x17, 0x0a, 0x13, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x5f, 0x54, 0x4f,
	0x5f, 0x45, 0x53, 0x54, 0x41, 0x42, 0x4c, 0x49, 0x53, 0x48, 0x10, 0x0a, 0x12, 0x18, 0x0a, 0x14,
	0x55, 0x4e, 0x53, 0x55, 0x50, 0x50, 0x4f, 0x52, 0x54, 0x45, 0x44, 0x5f, 0x57, 0x4f, 0x52, 0x4b,
	0x4c, 0x4f, 0x41, 0x44, 0x10, 0x0b, 0x12, 0x1a, 0x0a, 0x16, 0x4d, 0x49, 0x53, 0x43, 0x4f, 0x4e,
	0x46, 0x49, 0x47, 0x55, 0x52, 0x45, 0x44, 0x5f, 0x57, 0x4f, 0x52, 0x4b, 0x4c, 0x4f, 0x41, 0x44,
	0x10, 0x0e, 0x12, 0x0d, 0x0a, 0x09, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x10,
	0x0c, 0x12, 0x14, 0x0a, 0x10, 0x4d, 0x4f, 0x55, 0x4e, 0x54, 0x5f, 0x50, 0x4f, 0x49, 0x4e, 0x54,
	0x5f, 0x42, 0x55, 0x53, 0x59, 0x10, 0x0d, 0x12, 0x10, 0x0a, 0x0c, 0x55, 0x4e, 0x4b, 0x4e, 0x4f,
	0x57, 0x4e, 0x5f, 0x46, 0x4c, 0x41, 0x47, 0x10, 0x0f, 0x12, 0x0c, 0x0a, 0x08, 0x45, 0x58, 0x45,
	0x43, 0x5f, 0x43, 0x4d, 0x44, 0x10, 0x10, 0x42, 0x36, 0x5a, 0x34, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x72, 0x65, 0x73, 0x65, 0x6e,
	0x63, 0x65, 0x69, 0x6f, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x72, 0x65, 0x73, 0x65, 0x6e, 0x63,
	0x65, 0x2f, 0x72, 0x70, 0x63, 0x2f, 0x76, 0x32, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_common_errors_proto_rawDescOnce sync.Once
	file_common_errors_proto_rawDescData = file_common_errors_proto_rawDesc
)

func file_common_errors_proto_rawDescGZIP() []byte {
	file_common_errors_proto_rawDescOnce.Do(func() {
		file_common_errors_proto_rawDescData = protoimpl.X.CompressGZIP(file_common_errors_proto_rawDescData)
	})
	return file_common_errors_proto_rawDescData
}

var file_common_errors_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_common_errors_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_common_errors_proto_goTypes = []any{
	(InterceptError)(0),       // 0: telepresence.common.InterceptError
	(Result_ErrorCategory)(0), // 1: telepresence.common.Result.ErrorCategory
	(*Result)(nil),            // 2: telepresence.common.Result
}
var file_common_errors_proto_depIdxs = []int32{
	1, // 0: telepresence.common.Result.error_category:type_name -> telepresence.common.Result.ErrorCategory
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_common_errors_proto_init() }
func file_common_errors_proto_init() {
	if File_common_errors_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_common_errors_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_errors_proto_goTypes,
		DependencyIndexes: file_common_errors_proto_depIdxs,
		EnumInfos:         file_common_errors_proto_enumTypes,
		MessageInfos:      file_common_errors_proto_msgTypes,
	}.Build()
	File_common_errors_proto = out.File
	file_common_errors_proto_rawDesc = nil
	file_common_errors_proto_goTypes = nil
	file_common_errors_proto_depIdxs = nil
}
