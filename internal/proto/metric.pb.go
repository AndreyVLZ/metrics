// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.12.4
// source: internal/proto/metric.proto

package proto

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

type Info_Type int32

const (
	Info_COUNTER Info_Type = 0 // counter
	Info_GAUGE   Info_Type = 1 // gauge
)

// Enum value maps for Info_Type.
var (
	Info_Type_name = map[int32]string{
		0: "COUNTER",
		1: "GAUGE",
	}
	Info_Type_value = map[string]int32{
		"COUNTER": 0,
		"GAUGE":   1,
	}
)

func (x Info_Type) Enum() *Info_Type {
	p := new(Info_Type)
	*p = x
	return p
}

func (x Info_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Info_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_internal_proto_metric_proto_enumTypes[0].Descriptor()
}

func (Info_Type) Type() protoreflect.EnumType {
	return &file_internal_proto_metric_proto_enumTypes[0]
}

func (x Info_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Info_Type.Descriptor instead.
func (Info_Type) EnumDescriptor() ([]byte, []int) {
	return file_internal_proto_metric_proto_rawDescGZIP(), []int{0, 0}
}

type Info struct {
	state         protoimpl.MessageState
	Name          string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
	Type          Info_Type `protobuf:"varint,2,opt,name=type,proto3,enum=mymetric.Info_Type" json:"type,omitempty"`
}

func (x *Info) Reset() {
	*x = Info{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_metric_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Info) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Info) ProtoMessage() {}

func (x *Info) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_metric_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Info.ProtoReflect.Descriptor instead.
func (*Info) Descriptor() ([]byte, []int) {
	return file_internal_proto_metric_proto_rawDescGZIP(), []int{0}
}

func (x *Info) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Info) GetType() Info_Type {
	if x != nil {
		return x.Type
	}
	return Info_COUNTER
}

type Metric struct {
	state         protoimpl.MessageState
	Info          *Info `protobuf:"bytes,1,opt,name=info,proto3" json:"info,omitempty"`
	unknownFields protoimpl.UnknownFields
	Delta         int64   `protobuf:"varint,2,opt,name=delta,proto3" json:"delta,omitempty"`
	Value         float64 `protobuf:"fixed64,3,opt,name=value,proto3" json:"value,omitempty"`
	sizeCache     protoimpl.SizeCache
}

func (x *Metric) Reset() {
	*x = Metric{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_metric_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Metric) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metric) ProtoMessage() {}

func (x *Metric) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_metric_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Metric.ProtoReflect.Descriptor instead.
func (*Metric) Descriptor() ([]byte, []int) {
	return file_internal_proto_metric_proto_rawDescGZIP(), []int{1}
}

func (x *Metric) GetInfo() *Info {
	if x != nil {
		return x.Info
	}
	return nil
}

func (x *Metric) GetDelta() int64 {
	if x != nil {
		return x.Delta
	}
	return 0
}

func (x *Metric) GetValue() float64 {
	if x != nil {
		return x.Value
	}
	return 0
}

// Добавление списка метрик [запрос].
type AddBatchRequest struct {
	state         protoimpl.MessageState
	unknownFields protoimpl.UnknownFields
	Arr           []*Metric `protobuf:"bytes,1,rep,name=arr,proto3" json:"arr,omitempty"`
	sizeCache     protoimpl.SizeCache
}

func (x *AddBatchRequest) Reset() {
	*x = AddBatchRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_metric_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddBatchRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddBatchRequest) ProtoMessage() {}

func (x *AddBatchRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_metric_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddBatchRequest.ProtoReflect.Descriptor instead.
func (*AddBatchRequest) Descriptor() ([]byte, []int) {
	return file_internal_proto_metric_proto_rawDescGZIP(), []int{2}
}

func (x *AddBatchRequest) GetArr() []*Metric {
	if x != nil {
		return x.Arr
	}
	return nil
}

// Добавление списка метрик [ответ].
type AddBatchResponse struct {
	state         protoimpl.MessageState
	Error         string `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AddBatchResponse) Reset() {
	*x = AddBatchResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_metric_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddBatchResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddBatchResponse) ProtoMessage() {}

func (x *AddBatchResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_metric_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddBatchResponse.ProtoReflect.Descriptor instead.
func (*AddBatchResponse) Descriptor() ([]byte, []int) {
	return file_internal_proto_metric_proto_rawDescGZIP(), []int{3}
}

func (x *AddBatchResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_internal_proto_metric_proto protoreflect.FileDescriptor

var file_internal_proto_metric_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x6d,
	0x79, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x22, 0x63, 0x0a, 0x04, 0x49, 0x6e, 0x66, 0x6f, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x27, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x13, 0x2e, 0x6d, 0x79, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x2e, 0x49, 0x6e, 0x66,
	0x6f, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22, 0x1e, 0x0a, 0x04,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x43, 0x4f, 0x55, 0x4e, 0x54, 0x45, 0x52, 0x10,
	0x00, 0x12, 0x09, 0x0a, 0x05, 0x47, 0x41, 0x55, 0x47, 0x45, 0x10, 0x01, 0x22, 0x58, 0x0a, 0x06,
	0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x12, 0x22, 0x0a, 0x04, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x6d, 0x79, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x2e,
	0x49, 0x6e, 0x66, 0x6f, 0x52, 0x04, 0x69, 0x6e, 0x66, 0x6f, 0x12, 0x14, 0x0a, 0x05, 0x64, 0x65,
	0x6c, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x64, 0x65, 0x6c, 0x74, 0x61,
	0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x35, 0x0a, 0x0f, 0x41, 0x64, 0x64, 0x42, 0x61, 0x74,
	0x63, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x22, 0x0a, 0x03, 0x61, 0x72, 0x72,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x6d, 0x79, 0x6d, 0x65, 0x74, 0x72, 0x69,
	0x63, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x03, 0x61, 0x72, 0x72, 0x22, 0x28, 0x0a,
	0x10, 0x41, 0x64, 0x64, 0x42, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x32, 0x4c, 0x0a, 0x07, 0x4d, 0x65, 0x74, 0x72, 0x69,
	0x63, 0x73, 0x12, 0x41, 0x0a, 0x08, 0x41, 0x64, 0x64, 0x42, 0x61, 0x74, 0x63, 0x68, 0x12, 0x19,
	0x2e, 0x6d, 0x79, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x2e, 0x41, 0x64, 0x64, 0x42, 0x61, 0x74,
	0x63, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x6d, 0x79, 0x6d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x2e, 0x41, 0x64, 0x64, 0x42, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x10, 0x5a, 0x0e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_proto_metric_proto_rawDescOnce sync.Once
	file_internal_proto_metric_proto_rawDescData = file_internal_proto_metric_proto_rawDesc
)

func file_internal_proto_metric_proto_rawDescGZIP() []byte {
	file_internal_proto_metric_proto_rawDescOnce.Do(func() {
		file_internal_proto_metric_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_proto_metric_proto_rawDescData)
	})
	return file_internal_proto_metric_proto_rawDescData
}

var file_internal_proto_metric_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_internal_proto_metric_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_internal_proto_metric_proto_goTypes = []any{
	(Info_Type)(0),           // 0: mymetric.Info.Type
	(*Info)(nil),             // 1: mymetric.Info
	(*Metric)(nil),           // 2: mymetric.Metric
	(*AddBatchRequest)(nil),  // 3: mymetric.AddBatchRequest
	(*AddBatchResponse)(nil), // 4: mymetric.AddBatchResponse
}
var file_internal_proto_metric_proto_depIdxs = []int32{
	0, // 0: mymetric.Info.type:type_name -> mymetric.Info.Type
	1, // 1: mymetric.Metric.info:type_name -> mymetric.Info
	2, // 2: mymetric.AddBatchRequest.arr:type_name -> mymetric.Metric
	3, // 3: mymetric.Metrics.AddBatch:input_type -> mymetric.AddBatchRequest
	4, // 4: mymetric.Metrics.AddBatch:output_type -> mymetric.AddBatchResponse
	4, // [4:5] is the sub-list for method output_type
	3, // [3:4] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_internal_proto_metric_proto_init() }
func file_internal_proto_metric_proto_init() {
	if File_internal_proto_metric_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_proto_metric_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Info); i {
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
		file_internal_proto_metric_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*Metric); i {
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
		file_internal_proto_metric_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*AddBatchRequest); i {
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
		file_internal_proto_metric_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*AddBatchResponse); i {
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
			RawDescriptor: file_internal_proto_metric_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_internal_proto_metric_proto_goTypes,
		DependencyIndexes: file_internal_proto_metric_proto_depIdxs,
		EnumInfos:         file_internal_proto_metric_proto_enumTypes,
		MessageInfos:      file_internal_proto_metric_proto_msgTypes,
	}.Build()
	File_internal_proto_metric_proto = out.File
	file_internal_proto_metric_proto_rawDesc = nil
	file_internal_proto_metric_proto_goTypes = nil
	file_internal_proto_metric_proto_depIdxs = nil
}
