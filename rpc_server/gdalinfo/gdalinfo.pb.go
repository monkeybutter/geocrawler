// Code generated by protoc-gen-go.
// source: gdalinfo.proto
// DO NOT EDIT!

/*
Package gdalinfo is a generated protocol buffer package.

It is generated from these files:
	gdalinfo.proto

It has these top-level messages:
	Request
	Overview
	Pair
	Map
	GDALDataSet
	GDALFile
*/
package gdalinfo

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Request struct {
	FilePath string `protobuf:"bytes,1,opt,name=filePath" json:"filePath,omitempty"`
}

func (m *Request) Reset()                    { *m = Request{} }
func (m *Request) String() string            { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()               {}
func (*Request) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Request) GetFilePath() string {
	if m != nil {
		return m.FilePath
	}
	return ""
}

type Overview struct {
	XSize int32 `protobuf:"varint,1,opt,name=xSize" json:"xSize,omitempty"`
	YSize int32 `protobuf:"varint,2,opt,name=ySize" json:"ySize,omitempty"`
}

func (m *Overview) Reset()                    { *m = Overview{} }
func (m *Overview) String() string            { return proto.CompactTextString(m) }
func (*Overview) ProtoMessage()               {}
func (*Overview) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Overview) GetXSize() int32 {
	if m != nil {
		return m.XSize
	}
	return 0
}

func (m *Overview) GetYSize() int32 {
	if m != nil {
		return m.YSize
	}
	return 0
}

type Pair struct {
	Key   string `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
}

func (m *Pair) Reset()                    { *m = Pair{} }
func (m *Pair) String() string            { return proto.CompactTextString(m) }
func (*Pair) ProtoMessage()               {}
func (*Pair) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Pair) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *Pair) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type Map struct {
	Pairs []*Pair `protobuf:"bytes,1,rep,name=pairs" json:"pairs,omitempty"`
}

func (m *Map) Reset()                    { *m = Map{} }
func (m *Map) String() string            { return proto.CompactTextString(m) }
func (*Map) ProtoMessage()               {}
func (*Map) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Map) GetPairs() []*Pair {
	if m != nil {
		return m.Pairs
	}
	return nil
}

type GDALDataSet struct {
	DataSetName  string      `protobuf:"bytes,1,opt,name=dataSetName" json:"dataSetName,omitempty"`
	RasterCount  int32       `protobuf:"varint,2,opt,name=rasterCount" json:"rasterCount,omitempty"`
	Type         string      `protobuf:"bytes,3,opt,name=type" json:"type,omitempty"`
	XSize        int32       `protobuf:"varint,4,opt,name=xSize" json:"xSize,omitempty"`
	YSize        int32       `protobuf:"varint,5,opt,name=ySize" json:"ySize,omitempty"`
	ProjWKT      string      `protobuf:"bytes,6,opt,name=projWKT" json:"projWKT,omitempty"`
	GeoTransform []float64   `protobuf:"fixed64,7,rep,packed,name=geoTransform" json:"geoTransform,omitempty"`
	Overviews    []*Overview `protobuf:"bytes,8,rep,name=overviews" json:"overviews,omitempty"`
	Extras       *Map        `protobuf:"bytes,9,opt,name=extras" json:"extras,omitempty"`
}

func (m *GDALDataSet) Reset()                    { *m = GDALDataSet{} }
func (m *GDALDataSet) String() string            { return proto.CompactTextString(m) }
func (*GDALDataSet) ProtoMessage()               {}
func (*GDALDataSet) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *GDALDataSet) GetDataSetName() string {
	if m != nil {
		return m.DataSetName
	}
	return ""
}

func (m *GDALDataSet) GetRasterCount() int32 {
	if m != nil {
		return m.RasterCount
	}
	return 0
}

func (m *GDALDataSet) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *GDALDataSet) GetXSize() int32 {
	if m != nil {
		return m.XSize
	}
	return 0
}

func (m *GDALDataSet) GetYSize() int32 {
	if m != nil {
		return m.YSize
	}
	return 0
}

func (m *GDALDataSet) GetProjWKT() string {
	if m != nil {
		return m.ProjWKT
	}
	return ""
}

func (m *GDALDataSet) GetGeoTransform() []float64 {
	if m != nil {
		return m.GeoTransform
	}
	return nil
}

func (m *GDALDataSet) GetOverviews() []*Overview {
	if m != nil {
		return m.Overviews
	}
	return nil
}

func (m *GDALDataSet) GetExtras() *Map {
	if m != nil {
		return m.Extras
	}
	return nil
}

type GDALFile struct {
	FileName string         `protobuf:"bytes,1,opt,name=fileName" json:"fileName,omitempty"`
	Driver   string         `protobuf:"bytes,2,opt,name=driver" json:"driver,omitempty"`
	DataSets []*GDALDataSet `protobuf:"bytes,3,rep,name=dataSets" json:"dataSets,omitempty"`
}

func (m *GDALFile) Reset()                    { *m = GDALFile{} }
func (m *GDALFile) String() string            { return proto.CompactTextString(m) }
func (*GDALFile) ProtoMessage()               {}
func (*GDALFile) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *GDALFile) GetFileName() string {
	if m != nil {
		return m.FileName
	}
	return ""
}

func (m *GDALFile) GetDriver() string {
	if m != nil {
		return m.Driver
	}
	return ""
}

func (m *GDALFile) GetDataSets() []*GDALDataSet {
	if m != nil {
		return m.DataSets
	}
	return nil
}

func init() {
	proto.RegisterType((*Request)(nil), "gdalinfo.Request")
	proto.RegisterType((*Overview)(nil), "gdalinfo.Overview")
	proto.RegisterType((*Pair)(nil), "gdalinfo.Pair")
	proto.RegisterType((*Map)(nil), "gdalinfo.Map")
	proto.RegisterType((*GDALDataSet)(nil), "gdalinfo.GDALDataSet")
	proto.RegisterType((*GDALFile)(nil), "gdalinfo.GDALFile")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for GDALInfo service

type GDALInfoClient interface {
	Extract(ctx context.Context, in *Request, opts ...grpc.CallOption) (*GDALFile, error)
}

type gDALInfoClient struct {
	cc *grpc.ClientConn
}

func NewGDALInfoClient(cc *grpc.ClientConn) GDALInfoClient {
	return &gDALInfoClient{cc}
}

func (c *gDALInfoClient) Extract(ctx context.Context, in *Request, opts ...grpc.CallOption) (*GDALFile, error) {
	out := new(GDALFile)
	err := grpc.Invoke(ctx, "/gdalinfo.GDALInfo/Extract", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for GDALInfo service

type GDALInfoServer interface {
	Extract(context.Context, *Request) (*GDALFile, error)
}

func RegisterGDALInfoServer(s *grpc.Server, srv GDALInfoServer) {
	s.RegisterService(&_GDALInfo_serviceDesc, srv)
}

func _GDALInfo_Extract_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GDALInfoServer).Extract(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gdalinfo.GDALInfo/Extract",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GDALInfoServer).Extract(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

var _GDALInfo_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gdalinfo.GDALInfo",
	HandlerType: (*GDALInfoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Extract",
			Handler:    _GDALInfo_Extract_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "gdalinfo.proto",
}

func init() { proto.RegisterFile("gdalinfo.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 355 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x92, 0x5f, 0x4f, 0xe2, 0x40,
	0x14, 0xc5, 0xd3, 0x2d, 0xfd, 0x77, 0x0b, 0xec, 0xee, 0xb0, 0x9b, 0x4c, 0x34, 0x24, 0x4d, 0xa3,
	0xb1, 0x4f, 0xc4, 0xe0, 0xab, 0x2f, 0x46, 0xd4, 0x18, 0x45, 0x89, 0x60, 0x7c, 0x1e, 0xe5, 0x16,
	0x47, 0x4b, 0xa7, 0x4e, 0x87, 0x0a, 0x7e, 0x4a, 0x3f, 0x92, 0xe9, 0x1f, 0x28, 0xbc, 0xf5, 0xa4,
	0x67, 0xce, 0xfd, 0xdd, 0x33, 0x03, 0xed, 0xd9, 0x94, 0x45, 0x3c, 0x0e, 0x45, 0x2f, 0x91, 0x42,
	0x09, 0x62, 0xaf, 0xb5, 0xbf, 0x0f, 0xd6, 0x03, 0x7e, 0x2c, 0x30, 0x55, 0xe4, 0x0f, 0xd8, 0x21,
	0x8f, 0x70, 0xc4, 0xd4, 0x2b, 0xd5, 0x3c, 0x2d, 0x70, 0xfc, 0x00, 0xec, 0xfb, 0x0c, 0x65, 0xc6,
	0xf1, 0x93, 0xb4, 0xc0, 0x58, 0x8e, 0xf9, 0x17, 0x16, 0xbf, 0x8c, 0x5c, 0xae, 0x0a, 0xf9, 0x2b,
	0x97, 0xbe, 0x0f, 0x8d, 0x11, 0xe3, 0x92, 0xb8, 0xa0, 0xbf, 0xe3, 0xaa, 0x3c, 0x9e, 0x7b, 0x32,
	0x16, 0x2d, 0x4a, 0x8f, 0xe3, 0x1f, 0x80, 0x3e, 0x64, 0x09, 0xe9, 0x82, 0x91, 0x30, 0x2e, 0x53,
	0xaa, 0x79, 0x7a, 0xe0, 0xf6, 0xdb, 0xbd, 0x0d, 0x5b, 0x9e, 0xe0, 0x7f, 0x6b, 0xe0, 0x5e, 0x0d,
	0xce, 0x6e, 0x07, 0x4c, 0xb1, 0x31, 0x2a, 0xd2, 0x01, 0x77, 0x5a, 0x7e, 0xde, 0xb1, 0x39, 0x56,
	0xc9, 0x1d, 0x70, 0x25, 0x4b, 0x15, 0xca, 0x73, 0xb1, 0x88, 0x55, 0xc9, 0x40, 0x9a, 0xd0, 0x50,
	0xab, 0x04, 0xa9, 0xbe, 0x1e, 0x5e, 0xf2, 0x36, 0x76, 0x79, 0x8d, 0x42, 0xfe, 0x06, 0x2b, 0x91,
	0xe2, 0xed, 0xe9, 0x66, 0x42, 0xcd, 0xc2, 0xfe, 0x0f, 0x9a, 0x33, 0x14, 0x13, 0xc9, 0xe2, 0x34,
	0x14, 0x72, 0x4e, 0x2d, 0x4f, 0x0f, 0x34, 0x72, 0x08, 0x8e, 0xa8, 0x0a, 0x48, 0xa9, 0x5d, 0xf0,
	0x92, 0x9a, 0x77, 0xd3, 0x4d, 0x17, 0x4c, 0x5c, 0x2a, 0xc9, 0x52, 0xea, 0x78, 0x5a, 0xe0, 0xf6,
	0x5b, 0xb5, 0x67, 0xc8, 0x12, 0xff, 0x11, 0xec, 0x7c, 0xa3, 0x4b, 0x1e, 0xe1, 0xba, 0xe4, 0xad,
	0x5d, 0xda, 0x60, 0x4e, 0x25, 0xcf, 0x50, 0x96, 0x35, 0x91, 0x23, 0xb0, 0xab, 0x85, 0x53, 0xaa,
	0x17, 0x23, 0xff, 0xd7, 0x71, 0x5b, 0xcd, 0xf4, 0x4f, 0xcb, 0xd8, 0xeb, 0x38, 0x14, 0xe4, 0x18,
	0xac, 0x8b, 0x9c, 0xe0, 0x45, 0x91, 0xbf, 0xb5, 0xbb, 0xba, 0xd9, 0x3d, 0xb2, 0x1b, 0x90, 0x83,
	0x3c, 0x9b, 0xc5, 0x4b, 0x38, 0xf9, 0x09, 0x00, 0x00, 0xff, 0xff, 0x8f, 0x93, 0x97, 0x4b, 0x1b,
	0x02, 0x00, 0x00,
}