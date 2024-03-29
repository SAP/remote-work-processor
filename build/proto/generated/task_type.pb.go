// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.9
// source: task_type.proto

package pb

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

type TaskType int32

const (
	TaskType_TASK_TYPE_INVALID                TaskType = 0
	TaskType_TASK_TYPE_VOID                   TaskType = 1
	TaskType_TASK_TYPE_HTTP                   TaskType = 2
	TaskType_TASK_TYPE_SCRIPT                 TaskType = 3
	TaskType_TASK_TYPE_KUBERNETES_API_REQUEST TaskType = 4
)

// Enum value maps for TaskType.
var (
	TaskType_name = map[int32]string{
		0: "TASK_TYPE_INVALID",
		1: "TASK_TYPE_VOID",
		2: "TASK_TYPE_HTTP",
		3: "TASK_TYPE_SCRIPT",
		4: "TASK_TYPE_KUBERNETES_API_REQUEST",
	}
	TaskType_value = map[string]int32{
		"TASK_TYPE_INVALID":                0,
		"TASK_TYPE_VOID":                   1,
		"TASK_TYPE_HTTP":                   2,
		"TASK_TYPE_SCRIPT":                 3,
		"TASK_TYPE_KUBERNETES_API_REQUEST": 4,
	}
)

func (x TaskType) Enum() *TaskType {
	p := new(TaskType)
	*p = x
	return p
}

func (x TaskType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TaskType) Descriptor() protoreflect.EnumDescriptor {
	return file_task_type_proto_enumTypes[0].Descriptor()
}

func (TaskType) Type() protoreflect.EnumType {
	return &file_task_type_proto_enumTypes[0]
}

func (x TaskType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TaskType.Descriptor instead.
func (TaskType) EnumDescriptor() ([]byte, []int) {
	return file_task_type_proto_rawDescGZIP(), []int{0}
}

var File_task_type_proto protoreflect.FileDescriptor

var file_task_type_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x26, 0x73, 0x61, 0x70, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x70, 0x69, 0x6c, 0x6f, 0x74,
	0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x70, 0x72, 0x6f,
	0x63, 0x65, 0x73, 0x73, 0x6f, 0x72, 0x2e, 0x76, 0x31, 0x2a, 0x85, 0x01, 0x0a, 0x08, 0x54, 0x61,
	0x73, 0x6b, 0x54, 0x79, 0x70, 0x65, 0x12, 0x15, 0x0a, 0x11, 0x54, 0x41, 0x53, 0x4b, 0x5f, 0x54,
	0x59, 0x50, 0x45, 0x5f, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x10, 0x00, 0x12, 0x12, 0x0a,
	0x0e, 0x54, 0x41, 0x53, 0x4b, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x56, 0x4f, 0x49, 0x44, 0x10,
	0x01, 0x12, 0x12, 0x0a, 0x0e, 0x54, 0x41, 0x53, 0x4b, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x48,
	0x54, 0x54, 0x50, 0x10, 0x02, 0x12, 0x14, 0x0a, 0x10, 0x54, 0x41, 0x53, 0x4b, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x53, 0x43, 0x52, 0x49, 0x50, 0x54, 0x10, 0x03, 0x12, 0x24, 0x0a, 0x20, 0x54,
	0x41, 0x53, 0x4b, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x4b, 0x55, 0x42, 0x45, 0x52, 0x4e, 0x45,
	0x54, 0x45, 0x53, 0x5f, 0x41, 0x50, 0x49, 0x5f, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x10,
	0x04, 0x42, 0x76, 0x0a, 0x30, 0x63, 0x6f, 0x6d, 0x2e, 0x73, 0x61, 0x70, 0x2e, 0x61, 0x75, 0x74,
	0x6f, 0x70, 0x69, 0x6c, 0x6f, 0x74, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2e, 0x77, 0x6f,
	0x72, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x42, 0x0d, 0x54, 0x61, 0x73, 0x6b, 0x54, 0x79, 0x70, 0x65, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x50, 0x00, 0x5a, 0x31, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x53, 0x41, 0x50, 0x2f, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2d, 0x77, 0x6f,
	0x72, 0x6b, 0x2d, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x6f, 0x72, 0x2f, 0x67, 0x65, 0x6e,
	0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x3b, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_task_type_proto_rawDescOnce sync.Once
	file_task_type_proto_rawDescData = file_task_type_proto_rawDesc
)

func file_task_type_proto_rawDescGZIP() []byte {
	file_task_type_proto_rawDescOnce.Do(func() {
		file_task_type_proto_rawDescData = protoimpl.X.CompressGZIP(file_task_type_proto_rawDescData)
	})
	return file_task_type_proto_rawDescData
}

var file_task_type_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_task_type_proto_goTypes = []interface{}{
	(TaskType)(0), // 0: sap.autopilot.remote.work.processor.v1.TaskType
}
var file_task_type_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_task_type_proto_init() }
func file_task_type_proto_init() {
	if File_task_type_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_task_type_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_task_type_proto_goTypes,
		DependencyIndexes: file_task_type_proto_depIdxs,
		EnumInfos:         file_task_type_proto_enumTypes,
	}.Build()
	File_task_type_proto = out.File
	file_task_type_proto_rawDesc = nil
	file_task_type_proto_goTypes = nil
	file_task_type_proto_depIdxs = nil
}
