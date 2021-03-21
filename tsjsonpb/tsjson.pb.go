// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.10.1
// source: tsjson.proto

package tsjsonpb

import (
	proto "github.com/golang/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

var file_tsjson_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptor.FileOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         210320,
		Name:          "tsjson.npm_package",
		Tag:           "bytes,210320,opt,name=npm_package",
		Filename:      "tsjson.proto",
	},
	{
		ExtendedType:  (*descriptor.FileOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         210321,
		Name:          "tsjson.import_path",
		Tag:           "bytes,210321,opt,name=import_path",
		Filename:      "tsjson.proto",
	},
}

// Extension fields to descriptor.FileOptions.
var (
	// Specifies an NPM package to use as an import statement when traversing packages
	//
	// optional string npm_package = 210320;
	E_NpmPackage = &file_tsjson_proto_extTypes[0]
	// Specifies the path from the root of the package to the file
	//
	// optional string import_path = 210321;
	E_ImportPath = &file_tsjson_proto_extTypes[1]
)

var File_tsjson_proto protoreflect.FileDescriptor

var file_tsjson_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x74, 0x73, 0x6a, 0x73, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x74, 0x73, 0x6a, 0x73, 0x6f, 0x6e, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x3a, 0x3f, 0x0a, 0x0b, 0x6e, 0x70, 0x6d, 0x5f,
	0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x12, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x4f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x90, 0xeb, 0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6e,
	0x70, 0x6d, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x3a, 0x3f, 0x0a, 0x0b, 0x69, 0x6d, 0x70,
	0x6f, 0x72, 0x74, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x12, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x4f,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x91, 0xeb, 0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x69, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x50, 0x61, 0x74, 0x68, 0x42, 0x51, 0x5a, 0x2f, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x4c, 0x4c, 0x4b, 0x65, 0x6e, 0x6e, 0x65,
	0x64, 0x79, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x74, 0x73,
	0x6a, 0x73, 0x6f, 0x6e, 0x2f, 0x74, 0x73, 0x6a, 0x73, 0x6f, 0x6e, 0x70, 0x62, 0x82, 0xd9, 0x66,
	0x1c, 0x40, 0x6c, 0x6c, 0x6b, 0x65, 0x6e, 0x6e, 0x65, 0x64, 0x79, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x74, 0x73, 0x6a, 0x73, 0x6f, 0x6e, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_tsjson_proto_goTypes = []interface{}{
	(*descriptor.FileOptions)(nil), // 0: google.protobuf.FileOptions
}
var file_tsjson_proto_depIdxs = []int32{
	0, // 0: tsjson.npm_package:extendee -> google.protobuf.FileOptions
	0, // 1: tsjson.import_path:extendee -> google.protobuf.FileOptions
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	0, // [0:2] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_tsjson_proto_init() }
func file_tsjson_proto_init() {
	if File_tsjson_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_tsjson_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 2,
			NumServices:   0,
		},
		GoTypes:           file_tsjson_proto_goTypes,
		DependencyIndexes: file_tsjson_proto_depIdxs,
		ExtensionInfos:    file_tsjson_proto_extTypes,
	}.Build()
	File_tsjson_proto = out.File
	file_tsjson_proto_rawDesc = nil
	file_tsjson_proto_goTypes = nil
	file_tsjson_proto_depIdxs = nil
}