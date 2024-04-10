// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.6.1
// source: pluginmanager/pm.proto

package pluginmanager

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

type RunRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PluginsLibrary string `protobuf:"bytes,1,opt,name=PluginsLibrary,proto3" json:"PluginsLibrary,omitempty"`
	// oneof {
	Type string `protobuf:"bytes,2,opt,name=Type,proto3" json:"Type,omitempty"`
}

func (x *RunRequest) Reset() {
	*x = RunRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pluginmanager_pm_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RunRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunRequest) ProtoMessage() {}

func (x *RunRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pluginmanager_pm_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunRequest.ProtoReflect.Descriptor instead.
func (*RunRequest) Descriptor() ([]byte, []int) {
	return file_pluginmanager_pm_proto_rawDescGZIP(), []int{0}
}

func (x *RunRequest) GetPluginsLibrary() string {
	if x != nil {
		return x.PluginsLibrary
	}
	return ""
}

func (x *RunRequest) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

// RunResponse is the pm run status.
type RunResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type          string          `protobuf:"bytes,1,opt,name=Type,proto3" json:"Type,omitempty"`
	PluginsStatus []*PluginStatus `protobuf:"bytes,2,rep,name=PluginsStatus,proto3" json:"PluginsStatus,omitempty"`
	Status        string          `protobuf:"bytes,3,opt,name=Status,proto3" json:"Status,omitempty"`
	StdOutErr     string          `protobuf:"bytes,4,opt,name=StdOutErr,proto3" json:"StdOutErr,omitempty"`
}

func (x *RunResponse) Reset() {
	*x = RunResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pluginmanager_pm_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RunResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunResponse) ProtoMessage() {}

func (x *RunResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pluginmanager_pm_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunResponse.ProtoReflect.Descriptor instead.
func (*RunResponse) Descriptor() ([]byte, []int) {
	return file_pluginmanager_pm_proto_rawDescGZIP(), []int{1}
}

func (x *RunResponse) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *RunResponse) GetPluginsStatus() []*PluginStatus {
	if x != nil {
		return x.PluginsStatus
	}
	return nil
}

func (x *RunResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *RunResponse) GetStdOutErr() string {
	if x != nil {
		return x.StdOutErr
	}
	return ""
}

var File_pluginmanager_pm_proto protoreflect.FileDescriptor

var file_pluginmanager_pm_proto_rawDesc = []byte{
	0x0a, 0x16, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2f,
	0x70, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x1a, 0x1a, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x6d,
	0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x48, 0x0a, 0x0a, 0x52, 0x75, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x26, 0x0a, 0x0e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x4c, 0x69, 0x62, 0x72,
	0x61, 0x72, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x50, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x73, 0x4c, 0x69, 0x62, 0x72, 0x61, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x79, 0x70,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x54, 0x79, 0x70, 0x65, 0x22, 0x93, 0x01,
	0x0a, 0x0b, 0x52, 0x75, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x3a, 0x0a, 0x0d, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x0d,
	0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x16, 0x0a,
	0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x53, 0x74, 0x64, 0x4f, 0x75, 0x74, 0x45,
	0x72, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x53, 0x74, 0x64, 0x4f, 0x75, 0x74,
	0x45, 0x72, 0x72, 0x32, 0x42, 0x0a, 0x02, 0x50, 0x4d, 0x12, 0x3c, 0x0a, 0x03, 0x52, 0x75, 0x6e,
	0x12, 0x19, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72,
	0x2e, 0x52, 0x75, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x70, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x52, 0x75, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x1d, 0x5a, 0x1b, 0x70, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x6d,
	0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pluginmanager_pm_proto_rawDescOnce sync.Once
	file_pluginmanager_pm_proto_rawDescData = file_pluginmanager_pm_proto_rawDesc
)

func file_pluginmanager_pm_proto_rawDescGZIP() []byte {
	file_pluginmanager_pm_proto_rawDescOnce.Do(func() {
		file_pluginmanager_pm_proto_rawDescData = protoimpl.X.CompressGZIP(file_pluginmanager_pm_proto_rawDescData)
	})
	return file_pluginmanager_pm_proto_rawDescData
}

var file_pluginmanager_pm_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_pluginmanager_pm_proto_goTypes = []interface{}{
	(*RunRequest)(nil),   // 0: pluginmanager.RunRequest
	(*RunResponse)(nil),  // 1: pluginmanager.RunResponse
	(*PluginStatus)(nil), // 2: plugin.PluginStatus
}
var file_pluginmanager_pm_proto_depIdxs = []int32{
	2, // 0: pluginmanager.RunResponse.PluginsStatus:type_name -> plugin.PluginStatus
	0, // 1: pluginmanager.PM.Run:input_type -> pluginmanager.RunRequest
	1, // 2: pluginmanager.PM.Run:output_type -> pluginmanager.RunResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_pluginmanager_pm_proto_init() }
func file_pluginmanager_pm_proto_init() {
	if File_pluginmanager_pm_proto != nil {
		return
	}
	file_pluginmanager_plugin_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_pluginmanager_pm_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RunRequest); i {
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
		file_pluginmanager_pm_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RunResponse); i {
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
			RawDescriptor: file_pluginmanager_pm_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pluginmanager_pm_proto_goTypes,
		DependencyIndexes: file_pluginmanager_pm_proto_depIdxs,
		MessageInfos:      file_pluginmanager_pm_proto_msgTypes,
	}.Build()
	File_pluginmanager_pm_proto = out.File
	file_pluginmanager_pm_proto_rawDesc = nil
	file_pluginmanager_pm_proto_goTypes = nil
	file_pluginmanager_pm_proto_depIdxs = nil
}
