// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v5.26.1
// source: plugin.proto

package pluginmanager

import (
	runtime "github.com/VeritasOS/plugin-manager/utils/runtime"
	status "github.com/VeritasOS/plugin-manager/utils/status"
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

// PluginAttributes that are supported in a plugin file.
type PluginAttributes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FileName    string   `protobuf:"bytes,1,opt,name=FileName,proto3" json:"FileName,omitempty"`
	Description string   `protobuf:"bytes,2,opt,name=Description,proto3" json:"Description,omitempty"`
	ExecStart   string   `protobuf:"bytes,3,opt,name=ExecStart,proto3" json:"ExecStart,omitempty"`
	RequiredBy  []string `protobuf:"bytes,4,rep,name=RequiredBy,proto3" json:"RequiredBy,omitempty"`
	Requires    []string `protobuf:"bytes,5,rep,name=Requires,proto3" json:"Requires,omitempty"`
}

func (x *PluginAttributes) Reset() {
	*x = PluginAttributes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PluginAttributes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PluginAttributes) ProtoMessage() {}

func (x *PluginAttributes) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PluginAttributes.ProtoReflect.Descriptor instead.
func (*PluginAttributes) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{0}
}

func (x *PluginAttributes) GetFileName() string {
	if x != nil {
		return x.FileName
	}
	return ""
}

func (x *PluginAttributes) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *PluginAttributes) GetExecStart() string {
	if x != nil {
		return x.ExecStart
	}
	return ""
}

func (x *PluginAttributes) GetRequiredBy() []string {
	if x != nil {
		return x.RequiredBy
	}
	return nil
}

func (x *PluginAttributes) GetRequires() []string {
	if x != nil {
		return x.Requires
	}
	return nil
}

// Plugins is basically a map of file and its contents.
type Plugins struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Attributes map[string]*PluginAttributes `protobuf:"bytes,1,rep,name=Attributes,proto3" json:"Attributes,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Plugins) Reset() {
	*x = Plugins{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Plugins) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Plugins) ProtoMessage() {}

func (x *Plugins) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Plugins.ProtoReflect.Descriptor instead.
func (*Plugins) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{1}
}

func (x *Plugins) GetAttributes() map[string]*PluginAttributes {
	if x != nil {
		return x.Attributes
	}
	return nil
}

// PluginStatus is the plugin run's info: status, stdouterr.
type PluginStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Attributes *PluginAttributes `protobuf:"bytes,1,opt,name=Attributes,proto3" json:"Attributes,omitempty"`
	Status     status.Status     `protobuf:"varint,2,opt,name=Status,proto3,enum=status.Status" json:"Status,omitempty"`
	StdOutErr  []string          `protobuf:"bytes,3,rep,name=StdOutErr,proto3" json:"StdOutErr,omitempty"`
	RunTime    *runtime.RunTime  `protobuf:"bytes,4,opt,name=RunTime,proto3" json:"RunTime,omitempty"`
}

func (x *PluginStatus) Reset() {
	*x = PluginStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PluginStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PluginStatus) ProtoMessage() {}

func (x *PluginStatus) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PluginStatus.ProtoReflect.Descriptor instead.
func (*PluginStatus) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{2}
}

func (x *PluginStatus) GetAttributes() *PluginAttributes {
	if x != nil {
		return x.Attributes
	}
	return nil
}

func (x *PluginStatus) GetStatus() status.Status {
	if x != nil {
		return x.Status
	}
	return status.Status(0)
}

func (x *PluginStatus) GetStdOutErr() []string {
	if x != nil {
		return x.StdOutErr
	}
	return nil
}

func (x *PluginStatus) GetRunTime() *runtime.RunTime {
	if x != nil {
		return x.RunTime
	}
	return nil
}

// PluginTypeStatus is the pm run status of given Type plugin.
type PluginTypeStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type      string           `protobuf:"bytes,1,opt,name=Type,proto3" json:"Type,omitempty"`
	Plugins   []*PluginStatus  `protobuf:"bytes,2,rep,name=Plugins,proto3" json:"Plugins,omitempty"`
	Status    status.Status    `protobuf:"varint,3,opt,name=Status,proto3,enum=status.Status" json:"Status,omitempty"`
	StdOutErr string           `protobuf:"bytes,4,opt,name=StdOutErr,proto3" json:"StdOutErr,omitempty"`
	RunTime   *runtime.RunTime `protobuf:"bytes,5,opt,name=RunTime,proto3" json:"RunTime,omitempty"` // TODO: Add Percentage to get no. of pending vs. completed run of plugins.
}

func (x *PluginTypeStatus) Reset() {
	*x = PluginTypeStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PluginTypeStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PluginTypeStatus) ProtoMessage() {}

func (x *PluginTypeStatus) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PluginTypeStatus.ProtoReflect.Descriptor instead.
func (*PluginTypeStatus) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{3}
}

func (x *PluginTypeStatus) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *PluginTypeStatus) GetPlugins() []*PluginStatus {
	if x != nil {
		return x.Plugins
	}
	return nil
}

func (x *PluginTypeStatus) GetStatus() status.Status {
	if x != nil {
		return x.Status
	}
	return status.Status(0)
}

func (x *PluginTypeStatus) GetStdOutErr() string {
	if x != nil {
		return x.StdOutErr
	}
	return ""
}

func (x *PluginTypeStatus) GetRunTime() *runtime.RunTime {
	if x != nil {
		return x.RunTime
	}
	return nil
}

var File_plugin_proto protoreflect.FileDescriptor

var file_plugin_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x1a, 0x0d, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0c, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0xaa, 0x01, 0x0a, 0x10, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x41, 0x74,
	0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x46, 0x69, 0x6c, 0x65,
	0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x46, 0x69, 0x6c, 0x65,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x44, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x45, 0x78, 0x65, 0x63, 0x53, 0x74,
	0x61, 0x72, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x45, 0x78, 0x65, 0x63, 0x53,
	0x74, 0x61, 0x72, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x52, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64,
	0x42, 0x79, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a, 0x52, 0x65, 0x71, 0x75, 0x69, 0x72,
	0x65, 0x64, 0x42, 0x79, 0x12, 0x1a, 0x0a, 0x08, 0x52, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x73,
	0x18, 0x05, 0x20, 0x03, 0x28, 0x09, 0x52, 0x08, 0x52, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x73,
	0x22, 0xa3, 0x01, 0x0a, 0x07, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x12, 0x3f, 0x0a, 0x0a,
	0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1f, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x73, 0x2e, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x0a, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x1a, 0x57, 0x0a,
	0x0f, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x2e, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x18, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xba, 0x01, 0x0a, 0x0c, 0x50, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x38, 0x0a, 0x0a, 0x41, 0x74, 0x74, 0x72, 0x69,
	0x62, 0x75, 0x74, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x70, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x41, 0x74, 0x74, 0x72, 0x69,
	0x62, 0x75, 0x74, 0x65, 0x73, 0x52, 0x0a, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
	0x73, 0x12, 0x26, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x0e, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x52, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x53, 0x74, 0x64,
	0x4f, 0x75, 0x74, 0x45, 0x72, 0x72, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x53, 0x74,
	0x64, 0x4f, 0x75, 0x74, 0x45, 0x72, 0x72, 0x12, 0x2a, 0x0a, 0x07, 0x52, 0x75, 0x6e, 0x54, 0x69,
	0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69,
	0x6d, 0x65, 0x2e, 0x52, 0x75, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x52, 0x07, 0x52, 0x75, 0x6e, 0x54,
	0x69, 0x6d, 0x65, 0x22, 0xc8, 0x01, 0x0a, 0x10, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x54, 0x79,
	0x70, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x2e, 0x0a, 0x07,
	0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e,
	0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x07, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x12, 0x26, 0x0a, 0x06,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0e, 0x2e, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x53, 0x74, 0x64, 0x4f, 0x75, 0x74, 0x45, 0x72,
	0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x53, 0x74, 0x64, 0x4f, 0x75, 0x74, 0x45,
	0x72, 0x72, 0x12, 0x2a, 0x0a, 0x07, 0x52, 0x75, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x52, 0x75,
	0x6e, 0x54, 0x69, 0x6d, 0x65, 0x52, 0x07, 0x52, 0x75, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x42, 0x33,
	0x5a, 0x31, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x56, 0x65, 0x72,
	0x69, 0x74, 0x61, 0x73, 0x4f, 0x53, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2d, 0x6d, 0x61,
	0x6e, 0x61, 0x67, 0x65, 0x72, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x6d, 0x61, 0x6e, 0x61,
	0x67, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_plugin_proto_rawDescOnce sync.Once
	file_plugin_proto_rawDescData = file_plugin_proto_rawDesc
)

func file_plugin_proto_rawDescGZIP() []byte {
	file_plugin_proto_rawDescOnce.Do(func() {
		file_plugin_proto_rawDescData = protoimpl.X.CompressGZIP(file_plugin_proto_rawDescData)
	})
	return file_plugin_proto_rawDescData
}

var file_plugin_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_plugin_proto_goTypes = []interface{}{
	(*PluginAttributes)(nil), // 0: plugin.PluginAttributes
	(*Plugins)(nil),          // 1: plugin.Plugins
	(*PluginStatus)(nil),     // 2: plugin.PluginStatus
	(*PluginTypeStatus)(nil), // 3: plugin.PluginTypeStatus
	nil,                      // 4: plugin.Plugins.AttributesEntry
	(status.Status)(0),       // 5: status.Status
	(*runtime.RunTime)(nil),  // 6: runtime.RunTime
}
var file_plugin_proto_depIdxs = []int32{
	4, // 0: plugin.Plugins.Attributes:type_name -> plugin.Plugins.AttributesEntry
	0, // 1: plugin.PluginStatus.Attributes:type_name -> plugin.PluginAttributes
	5, // 2: plugin.PluginStatus.Status:type_name -> status.Status
	6, // 3: plugin.PluginStatus.RunTime:type_name -> runtime.RunTime
	2, // 4: plugin.PluginTypeStatus.Plugins:type_name -> plugin.PluginStatus
	5, // 5: plugin.PluginTypeStatus.Status:type_name -> status.Status
	6, // 6: plugin.PluginTypeStatus.RunTime:type_name -> runtime.RunTime
	0, // 7: plugin.Plugins.AttributesEntry.value:type_name -> plugin.PluginAttributes
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_plugin_proto_init() }
func file_plugin_proto_init() {
	if File_plugin_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_plugin_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PluginAttributes); i {
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
		file_plugin_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Plugins); i {
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
		file_plugin_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PluginStatus); i {
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
		file_plugin_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PluginTypeStatus); i {
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
			RawDescriptor: file_plugin_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_plugin_proto_goTypes,
		DependencyIndexes: file_plugin_proto_depIdxs,
		MessageInfos:      file_plugin_proto_msgTypes,
	}.Build()
	File_plugin_proto = out.File
	file_plugin_proto_rawDesc = nil
	file_plugin_proto_goTypes = nil
	file_plugin_proto_depIdxs = nil
}
