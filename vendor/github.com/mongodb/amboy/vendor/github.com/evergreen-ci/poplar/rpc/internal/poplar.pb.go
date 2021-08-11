// Code generated by protoc-gen-go. DO NOT EDIT.
// source: poplar.proto

package internal

import (
	fmt "fmt"
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

type CreateOptions_RecorderType int32

const (
	CreateOptions_UNKNOWN_RECORDER       CreateOptions_RecorderType = 0
	CreateOptions_PERF                   CreateOptions_RecorderType = 1
	CreateOptions_PERF_SINGLE            CreateOptions_RecorderType = 2
	CreateOptions_PERF_100MS             CreateOptions_RecorderType = 3
	CreateOptions_PERF_1S                CreateOptions_RecorderType = 4
	CreateOptions_HISTOGRAM_SINGLE       CreateOptions_RecorderType = 6
	CreateOptions_HISTOGRAM_100MS        CreateOptions_RecorderType = 7
	CreateOptions_HISTOGRAM_1S           CreateOptions_RecorderType = 8
	CreateOptions_INTERVAL_SUMMARIZATION CreateOptions_RecorderType = 9
)

var CreateOptions_RecorderType_name = map[int32]string{
	0: "UNKNOWN_RECORDER",
	1: "PERF",
	2: "PERF_SINGLE",
	3: "PERF_100MS",
	4: "PERF_1S",
	6: "HISTOGRAM_SINGLE",
	7: "HISTOGRAM_100MS",
	8: "HISTOGRAM_1S",
	9: "INTERVAL_SUMMARIZATION",
}

var CreateOptions_RecorderType_value = map[string]int32{
	"UNKNOWN_RECORDER":       0,
	"PERF":                   1,
	"PERF_SINGLE":            2,
	"PERF_100MS":             3,
	"PERF_1S":                4,
	"HISTOGRAM_SINGLE":       6,
	"HISTOGRAM_100MS":        7,
	"HISTOGRAM_1S":           8,
	"INTERVAL_SUMMARIZATION": 9,
}

func (x CreateOptions_RecorderType) String() string {
	return proto.EnumName(CreateOptions_RecorderType_name, int32(x))
}

func (CreateOptions_RecorderType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_b63ae76ef0e442c8, []int{1, 0}
}

type CreateOptions_EventsCollectorType int32

const (
	CreateOptions_UNKNOWN_COLLECTOR CreateOptions_EventsCollectorType = 0
	CreateOptions_BASIC             CreateOptions_EventsCollectorType = 1
	CreateOptions_PASSTHROUGH       CreateOptions_EventsCollectorType = 2
	CreateOptions_SAMPLING_100      CreateOptions_EventsCollectorType = 3
	CreateOptions_SAMPLING_1K       CreateOptions_EventsCollectorType = 4
	CreateOptions_SAMPLING_10K      CreateOptions_EventsCollectorType = 5
	CreateOptions_SAMPLING_100K     CreateOptions_EventsCollectorType = 6
	CreateOptions_RAND_SAMPLING_50  CreateOptions_EventsCollectorType = 7
	CreateOptions_RAND_SAMPLING_25  CreateOptions_EventsCollectorType = 8
	CreateOptions_RAND_SAMPLING_10  CreateOptions_EventsCollectorType = 9
	CreateOptions_INTERVAL_100MS    CreateOptions_EventsCollectorType = 10
	CreateOptions_INTERVAL_1S       CreateOptions_EventsCollectorType = 11
)

var CreateOptions_EventsCollectorType_name = map[int32]string{
	0:  "UNKNOWN_COLLECTOR",
	1:  "BASIC",
	2:  "PASSTHROUGH",
	3:  "SAMPLING_100",
	4:  "SAMPLING_1K",
	5:  "SAMPLING_10K",
	6:  "SAMPLING_100K",
	7:  "RAND_SAMPLING_50",
	8:  "RAND_SAMPLING_25",
	9:  "RAND_SAMPLING_10",
	10: "INTERVAL_100MS",
	11: "INTERVAL_1S",
}

var CreateOptions_EventsCollectorType_value = map[string]int32{
	"UNKNOWN_COLLECTOR": 0,
	"BASIC":             1,
	"PASSTHROUGH":       2,
	"SAMPLING_100":      3,
	"SAMPLING_1K":       4,
	"SAMPLING_10K":      5,
	"SAMPLING_100K":     6,
	"RAND_SAMPLING_50":  7,
	"RAND_SAMPLING_25":  8,
	"RAND_SAMPLING_10":  9,
	"INTERVAL_100MS":    10,
	"INTERVAL_1S":       11,
}

func (x CreateOptions_EventsCollectorType) String() string {
	return proto.EnumName(CreateOptions_EventsCollectorType_name, int32(x))
}

func (CreateOptions_EventsCollectorType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_b63ae76ef0e442c8, []int{1, 1}
}

type PoplarID struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PoplarID) Reset()         { *m = PoplarID{} }
func (m *PoplarID) String() string { return proto.CompactTextString(m) }
func (*PoplarID) ProtoMessage()    {}
func (*PoplarID) Descriptor() ([]byte, []int) {
	return fileDescriptor_b63ae76ef0e442c8, []int{0}
}

func (m *PoplarID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PoplarID.Unmarshal(m, b)
}
func (m *PoplarID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PoplarID.Marshal(b, m, deterministic)
}
func (m *PoplarID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PoplarID.Merge(m, src)
}
func (m *PoplarID) XXX_Size() int {
	return xxx_messageInfo_PoplarID.Size(m)
}
func (m *PoplarID) XXX_DiscardUnknown() {
	xxx_messageInfo_PoplarID.DiscardUnknown(m)
}

var xxx_messageInfo_PoplarID proto.InternalMessageInfo

func (m *PoplarID) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type CreateOptions struct {
	Name                 string                            `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Path                 string                            `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
	ChunkSize            int32                             `protobuf:"varint,3,opt,name=chunkSize,proto3" json:"chunkSize,omitempty"`
	Streaming            bool                              `protobuf:"varint,4,opt,name=streaming,proto3" json:"streaming,omitempty"`
	Dynamic              bool                              `protobuf:"varint,5,opt,name=dynamic,proto3" json:"dynamic,omitempty"`
	Recorder             CreateOptions_RecorderType        `protobuf:"varint,6,opt,name=recorder,proto3,enum=poplar.CreateOptions_RecorderType" json:"recorder,omitempty"`
	Events               CreateOptions_EventsCollectorType `protobuf:"varint,7,opt,name=events,proto3,enum=poplar.CreateOptions_EventsCollectorType" json:"events,omitempty"`
	Buffered             bool                              `protobuf:"varint,8,opt,name=buffered,proto3" json:"buffered,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                          `json:"-"`
	XXX_unrecognized     []byte                            `json:"-"`
	XXX_sizecache        int32                             `json:"-"`
}

func (m *CreateOptions) Reset()         { *m = CreateOptions{} }
func (m *CreateOptions) String() string { return proto.CompactTextString(m) }
func (*CreateOptions) ProtoMessage()    {}
func (*CreateOptions) Descriptor() ([]byte, []int) {
	return fileDescriptor_b63ae76ef0e442c8, []int{1}
}

func (m *CreateOptions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateOptions.Unmarshal(m, b)
}
func (m *CreateOptions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateOptions.Marshal(b, m, deterministic)
}
func (m *CreateOptions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateOptions.Merge(m, src)
}
func (m *CreateOptions) XXX_Size() int {
	return xxx_messageInfo_CreateOptions.Size(m)
}
func (m *CreateOptions) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateOptions.DiscardUnknown(m)
}

var xxx_messageInfo_CreateOptions proto.InternalMessageInfo

func (m *CreateOptions) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *CreateOptions) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *CreateOptions) GetChunkSize() int32 {
	if m != nil {
		return m.ChunkSize
	}
	return 0
}

func (m *CreateOptions) GetStreaming() bool {
	if m != nil {
		return m.Streaming
	}
	return false
}

func (m *CreateOptions) GetDynamic() bool {
	if m != nil {
		return m.Dynamic
	}
	return false
}

func (m *CreateOptions) GetRecorder() CreateOptions_RecorderType {
	if m != nil {
		return m.Recorder
	}
	return CreateOptions_UNKNOWN_RECORDER
}

func (m *CreateOptions) GetEvents() CreateOptions_EventsCollectorType {
	if m != nil {
		return m.Events
	}
	return CreateOptions_UNKNOWN_COLLECTOR
}

func (m *CreateOptions) GetBuffered() bool {
	if m != nil {
		return m.Buffered
	}
	return false
}

type PoplarResponse struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Status               bool     `protobuf:"varint,2,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PoplarResponse) Reset()         { *m = PoplarResponse{} }
func (m *PoplarResponse) String() string { return proto.CompactTextString(m) }
func (*PoplarResponse) ProtoMessage()    {}
func (*PoplarResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_b63ae76ef0e442c8, []int{2}
}

func (m *PoplarResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PoplarResponse.Unmarshal(m, b)
}
func (m *PoplarResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PoplarResponse.Marshal(b, m, deterministic)
}
func (m *PoplarResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PoplarResponse.Merge(m, src)
}
func (m *PoplarResponse) XXX_Size() int {
	return xxx_messageInfo_PoplarResponse.Size(m)
}
func (m *PoplarResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PoplarResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PoplarResponse proto.InternalMessageInfo

func (m *PoplarResponse) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *PoplarResponse) GetStatus() bool {
	if m != nil {
		return m.Status
	}
	return false
}

func init() {
	proto.RegisterEnum("poplar.CreateOptions_RecorderType", CreateOptions_RecorderType_name, CreateOptions_RecorderType_value)
	proto.RegisterEnum("poplar.CreateOptions_EventsCollectorType", CreateOptions_EventsCollectorType_name, CreateOptions_EventsCollectorType_value)
	proto.RegisterType((*PoplarID)(nil), "poplar.PoplarID")
	proto.RegisterType((*CreateOptions)(nil), "poplar.CreateOptions")
	proto.RegisterType((*PoplarResponse)(nil), "poplar.PoplarResponse")
}

func init() { proto.RegisterFile("poplar.proto", fileDescriptor_b63ae76ef0e442c8) }

var fileDescriptor_b63ae76ef0e442c8 = []byte{
	// 509 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x93, 0xd1, 0x6e, 0xda, 0x30,
	0x14, 0x86, 0x97, 0x02, 0x21, 0x1c, 0x28, 0x75, 0x4f, 0xb7, 0x2a, 0xaa, 0xa6, 0x09, 0xe5, 0x8a,
	0xdd, 0x20, 0xda, 0xa9, 0x77, 0xd3, 0xa4, 0x34, 0x64, 0x10, 0x01, 0x09, 0xb2, 0xc3, 0x26, 0xf5,
	0x06, 0xa5, 0xe0, 0xae, 0x68, 0x90, 0x44, 0x89, 0x99, 0xd4, 0xbd, 0xd7, 0x5e, 0x67, 0xaf, 0xb1,
	0xdb, 0xc9, 0x09, 0x10, 0xb6, 0x72, 0x77, 0xce, 0xe7, 0xf3, 0xdb, 0xff, 0x6f, 0xd9, 0xd0, 0x88,
	0xa3, 0x78, 0x15, 0x24, 0x9d, 0x38, 0x89, 0x44, 0x84, 0x6a, 0xde, 0x19, 0xef, 0x40, 0x9b, 0x64,
	0x95, 0xd3, 0x43, 0x84, 0x72, 0x18, 0xac, 0xb9, 0xae, 0xb4, 0x94, 0x76, 0x8d, 0x66, 0xb5, 0xf1,
	0xbb, 0x02, 0xa7, 0x56, 0xc2, 0x03, 0xc1, 0xbd, 0x58, 0x2c, 0xa3, 0x30, 0x3d, 0x36, 0x25, 0x59,
	0x1c, 0x88, 0x27, 0xfd, 0x24, 0x67, 0xb2, 0xc6, 0xb7, 0x50, 0x9b, 0x3f, 0x6d, 0xc2, 0xef, 0x6c,
	0xf9, 0x93, 0xeb, 0xa5, 0x96, 0xd2, 0xae, 0xd0, 0x02, 0xc8, 0xd5, 0x54, 0x24, 0x3c, 0x58, 0x2f,
	0xc3, 0x6f, 0x7a, 0xb9, 0xa5, 0xb4, 0x35, 0x5a, 0x00, 0xd4, 0xa1, 0xba, 0x78, 0x0e, 0x83, 0xf5,
	0x72, 0xae, 0x57, 0xb2, 0xb5, 0x5d, 0x8b, 0x9f, 0x40, 0x4b, 0xf8, 0x3c, 0x4a, 0x16, 0x3c, 0xd1,
	0xd5, 0x96, 0xd2, 0x6e, 0xde, 0x18, 0x9d, 0x6d, 0xb0, 0x7f, 0x6c, 0x76, 0xe8, 0x76, 0xca, 0x7f,
	0x8e, 0x39, 0xdd, 0x6b, 0xd0, 0x04, 0x95, 0xff, 0xe0, 0xa1, 0x48, 0xf5, 0x6a, 0xa6, 0x7e, 0x7f,
	0x5c, 0x6d, 0x67, 0x33, 0x56, 0xb4, 0x5a, 0xf1, 0xb9, 0x88, 0xf2, 0x4d, 0xb6, 0x42, 0xbc, 0x02,
	0xed, 0x61, 0xf3, 0xf8, 0xc8, 0x13, 0xbe, 0xd0, 0xb5, 0xcc, 0xdd, 0xbe, 0x37, 0x7e, 0x29, 0xd0,
	0x38, 0x3c, 0x19, 0x5f, 0x03, 0x99, 0xba, 0x43, 0xd7, 0xfb, 0xea, 0xce, 0xa8, 0x6d, 0x79, 0xb4,
	0x67, 0x53, 0xf2, 0x0a, 0x35, 0x28, 0x4f, 0x6c, 0xfa, 0x99, 0x28, 0x78, 0x06, 0x75, 0x59, 0xcd,
	0x98, 0xe3, 0xf6, 0x47, 0x36, 0x39, 0xc1, 0x26, 0x40, 0x06, 0xae, 0xbb, 0xdd, 0x31, 0x23, 0x25,
	0xac, 0x43, 0x35, 0xef, 0x19, 0x29, 0xcb, 0xdd, 0x06, 0x0e, 0xf3, 0xbd, 0x3e, 0x35, 0xc7, 0x3b,
	0x89, 0x8a, 0x17, 0x70, 0x56, 0xd0, 0x5c, 0x57, 0x45, 0x02, 0x8d, 0x03, 0xc8, 0x88, 0x86, 0x57,
	0x70, 0xe9, 0xb8, 0xbe, 0x4d, 0xbf, 0x98, 0xa3, 0x19, 0x9b, 0x8e, 0xc7, 0x26, 0x75, 0xee, 0x4d,
	0xdf, 0xf1, 0x5c, 0x52, 0x33, 0xfe, 0x28, 0x70, 0x71, 0x24, 0x33, 0xbe, 0x81, 0xf3, 0x9d, 0x7d,
	0xcb, 0x1b, 0x8d, 0x6c, 0xcb, 0xf7, 0xa4, 0xff, 0x1a, 0x54, 0xee, 0x4c, 0xe6, 0x58, 0xdb, 0x00,
	0x26, 0x63, 0xfe, 0x80, 0x7a, 0xd3, 0xfe, 0x80, 0x9c, 0xc8, 0x83, 0x99, 0x39, 0x9e, 0x8c, 0x1c,
	0xb7, 0x2f, 0xcd, 0x90, 0x92, 0x1c, 0x29, 0xc8, 0x90, 0x94, 0xff, 0x1b, 0x19, 0x92, 0x0a, 0x9e,
	0xc3, 0xe9, 0xa1, 0x68, 0x48, 0x54, 0x99, 0x95, 0x9a, 0x6e, 0x6f, 0xb6, 0xe7, 0xb7, 0x5d, 0x52,
	0x7d, 0x49, 0x6f, 0x6e, 0x89, 0xf6, 0x92, 0x5e, 0x77, 0x49, 0x0d, 0x11, 0x9a, 0xfb, 0xc0, 0xf9,
	0xb5, 0x80, 0xf4, 0x52, 0x30, 0x46, 0xea, 0xc6, 0x47, 0x68, 0xe6, 0x1f, 0x80, 0xf2, 0x34, 0x8e,
	0xc2, 0x94, 0x1f, 0x7d, 0xe0, 0x97, 0xa0, 0xa6, 0x22, 0x10, 0x9b, 0x34, 0x7b, 0xe2, 0x1a, 0xdd,
	0x76, 0x77, 0x70, 0xaf, 0x2d, 0x43, 0xc1, 0x93, 0x30, 0x58, 0x3d, 0xa8, 0xd9, 0xcf, 0xfa, 0xf0,
	0x37, 0x00, 0x00, 0xff, 0xff, 0xc2, 0xbe, 0x72, 0xbc, 0x69, 0x03, 0x00, 0x00,
}