// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sedachain/pubkey/v1/query.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// QueryParamsRequest is the request type for the Query/Params RPC method.
type QueryParamsRequest struct {
}

func (m *QueryParamsRequest) Reset()         { *m = QueryParamsRequest{} }
func (m *QueryParamsRequest) String() string { return proto.CompactTextString(m) }
func (*QueryParamsRequest) ProtoMessage()    {}
func (*QueryParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ab5fa3182b3fb474, []int{0}
}
func (m *QueryParamsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryParamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryParamsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryParamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryParamsRequest.Merge(m, src)
}
func (m *QueryParamsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryParamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryParamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryParamsRequest proto.InternalMessageInfo

// QueryParamsResponse is the response type for the Query/Params RPC method.
type QueryParamsResponse struct {
	// params defines the parameters of the module.
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

func (m *QueryParamsResponse) Reset()         { *m = QueryParamsResponse{} }
func (m *QueryParamsResponse) String() string { return proto.CompactTextString(m) }
func (*QueryParamsResponse) ProtoMessage()    {}
func (*QueryParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ab5fa3182b3fb474, []int{1}
}
func (m *QueryParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryParamsResponse.Merge(m, src)
}
func (m *QueryParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryParamsResponse proto.InternalMessageInfo

func (m *QueryParamsResponse) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

// QueryValidatorKeysRequest is request type for the Query/ValidatorKeys
// RPC method.
type QueryValidatorKeysRequest struct {
	ValidatorAddr string `protobuf:"bytes,1,opt,name=validator_addr,json=validatorAddr,proto3" json:"validator_addr,omitempty"`
}

func (m *QueryValidatorKeysRequest) Reset()         { *m = QueryValidatorKeysRequest{} }
func (m *QueryValidatorKeysRequest) String() string { return proto.CompactTextString(m) }
func (*QueryValidatorKeysRequest) ProtoMessage()    {}
func (*QueryValidatorKeysRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ab5fa3182b3fb474, []int{2}
}
func (m *QueryValidatorKeysRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryValidatorKeysRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryValidatorKeysRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryValidatorKeysRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryValidatorKeysRequest.Merge(m, src)
}
func (m *QueryValidatorKeysRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryValidatorKeysRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryValidatorKeysRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryValidatorKeysRequest proto.InternalMessageInfo

func (m *QueryValidatorKeysRequest) GetValidatorAddr() string {
	if m != nil {
		return m.ValidatorAddr
	}
	return ""
}

// QueryValidatorKeysResponse is response type for the Query/ValidatorKeys
// RPC method.
type QueryValidatorKeysResponse struct {
	ValidatorPubKeys ValidatorPubKeys `protobuf:"bytes,1,opt,name=validator_pub_keys,json=validatorPubKeys,proto3" json:"validator_pub_keys"`
}

func (m *QueryValidatorKeysResponse) Reset()         { *m = QueryValidatorKeysResponse{} }
func (m *QueryValidatorKeysResponse) String() string { return proto.CompactTextString(m) }
func (*QueryValidatorKeysResponse) ProtoMessage()    {}
func (*QueryValidatorKeysResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ab5fa3182b3fb474, []int{3}
}
func (m *QueryValidatorKeysResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryValidatorKeysResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryValidatorKeysResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryValidatorKeysResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryValidatorKeysResponse.Merge(m, src)
}
func (m *QueryValidatorKeysResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryValidatorKeysResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryValidatorKeysResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryValidatorKeysResponse proto.InternalMessageInfo

func (m *QueryValidatorKeysResponse) GetValidatorPubKeys() ValidatorPubKeys {
	if m != nil {
		return m.ValidatorPubKeys
	}
	return ValidatorPubKeys{}
}

// QueryProvingSchemesRequest is request type for the Query/ProvingSchemes
// RPC method.
type QueryProvingSchemesRequest struct {
}

func (m *QueryProvingSchemesRequest) Reset()         { *m = QueryProvingSchemesRequest{} }
func (m *QueryProvingSchemesRequest) String() string { return proto.CompactTextString(m) }
func (*QueryProvingSchemesRequest) ProtoMessage()    {}
func (*QueryProvingSchemesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ab5fa3182b3fb474, []int{4}
}
func (m *QueryProvingSchemesRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryProvingSchemesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryProvingSchemesRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryProvingSchemesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryProvingSchemesRequest.Merge(m, src)
}
func (m *QueryProvingSchemesRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryProvingSchemesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryProvingSchemesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryProvingSchemesRequest proto.InternalMessageInfo

// QueryProvingSchemesResponse is response type for the Query/ProvingSchemes
// RPC method.
type QueryProvingSchemesResponse struct {
	ProvingSchemes []ProvingScheme `protobuf:"bytes,1,rep,name=proving_schemes,json=provingSchemes,proto3" json:"proving_schemes"`
}

func (m *QueryProvingSchemesResponse) Reset()         { *m = QueryProvingSchemesResponse{} }
func (m *QueryProvingSchemesResponse) String() string { return proto.CompactTextString(m) }
func (*QueryProvingSchemesResponse) ProtoMessage()    {}
func (*QueryProvingSchemesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ab5fa3182b3fb474, []int{5}
}
func (m *QueryProvingSchemesResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryProvingSchemesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryProvingSchemesResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryProvingSchemesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryProvingSchemesResponse.Merge(m, src)
}
func (m *QueryProvingSchemesResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryProvingSchemesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryProvingSchemesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryProvingSchemesResponse proto.InternalMessageInfo

func (m *QueryProvingSchemesResponse) GetProvingSchemes() []ProvingScheme {
	if m != nil {
		return m.ProvingSchemes
	}
	return nil
}

func init() {
	proto.RegisterType((*QueryParamsRequest)(nil), "sedachain.pubkey.v1.QueryParamsRequest")
	proto.RegisterType((*QueryParamsResponse)(nil), "sedachain.pubkey.v1.QueryParamsResponse")
	proto.RegisterType((*QueryValidatorKeysRequest)(nil), "sedachain.pubkey.v1.QueryValidatorKeysRequest")
	proto.RegisterType((*QueryValidatorKeysResponse)(nil), "sedachain.pubkey.v1.QueryValidatorKeysResponse")
	proto.RegisterType((*QueryProvingSchemesRequest)(nil), "sedachain.pubkey.v1.QueryProvingSchemesRequest")
	proto.RegisterType((*QueryProvingSchemesResponse)(nil), "sedachain.pubkey.v1.QueryProvingSchemesResponse")
}

func init() { proto.RegisterFile("sedachain/pubkey/v1/query.proto", fileDescriptor_ab5fa3182b3fb474) }

var fileDescriptor_ab5fa3182b3fb474 = []byte{
	// 527 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0x8d, 0x29, 0x44, 0x62, 0xab, 0x06, 0xb4, 0xed, 0xa1, 0x71, 0x8a, 0x9b, 0x58, 0x20, 0x22,
	0xa4, 0x7a, 0x49, 0xe0, 0x02, 0x37, 0x72, 0x42, 0xea, 0x25, 0x4d, 0x25, 0x24, 0xb8, 0x44, 0x6b,
	0x7b, 0xe5, 0x58, 0x4d, 0xbc, 0x5b, 0xaf, 0x6d, 0xb0, 0x10, 0x1c, 0xf8, 0x02, 0x24, 0x3e, 0x80,
	0x1f, 0xe0, 0xc8, 0x47, 0xf4, 0x58, 0xc1, 0xa5, 0x27, 0x84, 0x12, 0x3e, 0x04, 0x79, 0x77, 0x9d,
	0xc6, 0xb0, 0xd0, 0xdc, 0xd6, 0x33, 0x6f, 0xde, 0xbc, 0x99, 0x37, 0x06, 0xfb, 0x9c, 0xf8, 0xd8,
	0x9b, 0xe0, 0x30, 0x42, 0x2c, 0x75, 0x4f, 0x48, 0x8e, 0xb2, 0x1e, 0x3a, 0x4d, 0x49, 0x9c, 0x3b,
	0x2c, 0xa6, 0x09, 0x85, 0xdb, 0x4b, 0x80, 0x23, 0x01, 0x4e, 0xd6, 0x33, 0xf7, 0x02, 0x4a, 0x83,
	0x29, 0x41, 0x98, 0x85, 0x08, 0x47, 0x11, 0x4d, 0x70, 0x12, 0xd2, 0x88, 0xcb, 0x12, 0x73, 0x27,
	0xa0, 0x01, 0x15, 0x4f, 0x54, 0xbc, 0x54, 0xb4, 0xe9, 0x51, 0x3e, 0xa3, 0x7c, 0x2c, 0x13, 0xf2,
	0x43, 0xa5, 0x3a, 0x3a, 0x11, 0x01, 0x89, 0x08, 0x0f, 0x4b, 0x48, 0x5b, 0x07, 0x51, 0x82, 0x04,
	0xc2, 0xde, 0x01, 0xf0, 0xa8, 0xd0, 0x3d, 0xc4, 0x31, 0x9e, 0xf1, 0x11, 0x39, 0x4d, 0x09, 0x4f,
	0xec, 0x21, 0xd8, 0xae, 0x44, 0x39, 0xa3, 0x11, 0x27, 0xf0, 0x09, 0xa8, 0x33, 0x11, 0xd9, 0x35,
	0xda, 0x46, 0x77, 0xb3, 0xdf, 0x72, 0x34, 0x63, 0x3a, 0xb2, 0x68, 0x70, 0xfd, 0xec, 0xc7, 0x7e,
	0x6d, 0xa4, 0x0a, 0x6c, 0x02, 0x9a, 0x82, 0xf1, 0x05, 0x9e, 0x86, 0x3e, 0x4e, 0x68, 0x7c, 0x48,
	0xf2, 0xb2, 0x1d, 0x7c, 0x0e, 0x1a, 0x59, 0x19, 0x1f, 0x63, 0xdf, 0x8f, 0x05, 0xff, 0xcd, 0x41,
	0xe7, 0xdb, 0xd7, 0x83, 0x3b, 0x6a, 0xe6, 0x65, 0xe1, 0x33, 0xdf, 0x8f, 0x09, 0xe7, 0xc7, 0x49,
	0x1c, 0x46, 0xc1, 0x68, 0x2b, 0x5b, 0x8d, 0xdb, 0xaf, 0x81, 0xa9, 0x6b, 0xa3, 0xf4, 0xbf, 0x04,
	0xf0, 0xb2, 0x0f, 0x4b, 0xdd, 0xf1, 0x09, 0xc9, 0xcb, 0x59, 0xee, 0x69, 0x67, 0x59, 0xf2, 0x0c,
	0x53, 0xb7, 0xa0, 0x52, 0x53, 0xdd, 0xce, 0xfe, 0x88, 0xdb, 0x7b, 0xaa, 0xf1, 0x30, 0xa6, 0x59,
	0x18, 0x05, 0xc7, 0xde, 0x84, 0xcc, 0xc8, 0x72, 0x9f, 0x0c, 0xb4, 0xb4, 0x59, 0xa5, 0xeb, 0x08,
	0xdc, 0x62, 0x32, 0x33, 0xe6, 0x32, 0xb5, 0x6b, 0xb4, 0x37, 0xba, 0x9b, 0x7d, 0x5b, 0xbf, 0xe0,
	0x55, 0x16, 0xa5, 0xa8, 0xc1, 0x2a, 0xd4, 0xfd, 0x8b, 0x0d, 0x70, 0x43, 0xb4, 0x84, 0xef, 0x41,
	0x5d, 0x3a, 0x02, 0xef, 0x6b, 0xd9, 0xfe, 0xb6, 0xdf, 0xec, 0x5e, 0x0d, 0x94, 0xca, 0xed, 0xce,
	0x87, 0xef, 0xbf, 0x3e, 0x5d, 0x6b, 0xc1, 0x26, 0x2a, 0x2a, 0x0e, 0x2a, 0xa7, 0x26, 0x9d, 0x87,
	0x5f, 0x0c, 0xb0, 0x55, 0xb1, 0x03, 0x3a, 0xff, 0xa6, 0xd7, 0x9d, 0x87, 0x89, 0xd6, 0xc6, 0x2b,
	0x55, 0x4f, 0x85, 0xaa, 0xc7, 0xb0, 0xaf, 0x51, 0x75, 0x79, 0x00, 0x85, 0xf9, 0xe8, 0x6d, 0xf5,
	0xf0, 0xde, 0xc1, 0xcf, 0x06, 0x68, 0x54, 0x6d, 0x82, 0xff, 0xe9, 0xaf, 0xb5, 0xdb, 0x7c, 0xb8,
	0x7e, 0x81, 0x52, 0xfc, 0x40, 0x28, 0xbe, 0x0b, 0x6d, 0xdd, 0x1e, 0xab, 0xa7, 0x31, 0x38, 0x3c,
	0x9b, 0x5b, 0xc6, 0xf9, 0xdc, 0x32, 0x7e, 0xce, 0x2d, 0xe3, 0xe3, 0xc2, 0xaa, 0x9d, 0x2f, 0xac,
	0xda, 0xc5, 0xc2, 0xaa, 0xbd, 0xea, 0x05, 0x61, 0x32, 0x49, 0x5d, 0xc7, 0xa3, 0x33, 0xc1, 0x23,
	0x7e, 0x71, 0x8f, 0x4e, 0x57, 0x49, 0xdf, 0x94, 0xb4, 0x49, 0xce, 0x08, 0x77, 0xeb, 0x02, 0xf3,
	0xe8, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0xe8, 0xae, 0xb5, 0x63, 0xd2, 0x04, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	// Params returns the total set of pubkey parameters.
	Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
	// ValidatorKeys returns a given validator's registered keys.
	ValidatorKeys(ctx context.Context, in *QueryValidatorKeysRequest, opts ...grpc.CallOption) (*QueryValidatorKeysResponse, error)
	// ProvingSchemes returns the statuses of the SEDA proving schemes.
	ProvingSchemes(ctx context.Context, in *QueryProvingSchemesRequest, opts ...grpc.CallOption) (*QueryProvingSchemesResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error) {
	out := new(QueryParamsResponse)
	err := c.cc.Invoke(ctx, "/sedachain.pubkey.v1.Query/Params", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ValidatorKeys(ctx context.Context, in *QueryValidatorKeysRequest, opts ...grpc.CallOption) (*QueryValidatorKeysResponse, error) {
	out := new(QueryValidatorKeysResponse)
	err := c.cc.Invoke(ctx, "/sedachain.pubkey.v1.Query/ValidatorKeys", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ProvingSchemes(ctx context.Context, in *QueryProvingSchemesRequest, opts ...grpc.CallOption) (*QueryProvingSchemesResponse, error) {
	out := new(QueryProvingSchemesResponse)
	err := c.cc.Invoke(ctx, "/sedachain.pubkey.v1.Query/ProvingSchemes", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// Params returns the total set of pubkey parameters.
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	// ValidatorKeys returns a given validator's registered keys.
	ValidatorKeys(context.Context, *QueryValidatorKeysRequest) (*QueryValidatorKeysResponse, error)
	// ProvingSchemes returns the statuses of the SEDA proving schemes.
	ProvingSchemes(context.Context, *QueryProvingSchemesRequest) (*QueryProvingSchemesResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Params not implemented")
}
func (*UnimplementedQueryServer) ValidatorKeys(ctx context.Context, req *QueryValidatorKeysRequest) (*QueryValidatorKeysResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidatorKeys not implemented")
}
func (*UnimplementedQueryServer) ProvingSchemes(ctx context.Context, req *QueryProvingSchemesRequest) (*QueryProvingSchemesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProvingSchemes not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Params(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sedachain.pubkey.v1.Query/Params",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ValidatorKeys_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryValidatorKeysRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ValidatorKeys(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sedachain.pubkey.v1.Query/ValidatorKeys",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ValidatorKeys(ctx, req.(*QueryValidatorKeysRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ProvingSchemes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryProvingSchemesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ProvingSchemes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sedachain.pubkey.v1.Query/ProvingSchemes",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ProvingSchemes(ctx, req.(*QueryProvingSchemesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc
var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "sedachain.pubkey.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Params",
			Handler:    _Query_Params_Handler,
		},
		{
			MethodName: "ValidatorKeys",
			Handler:    _Query_ValidatorKeys_Handler,
		},
		{
			MethodName: "ProvingSchemes",
			Handler:    _Query_ProvingSchemes_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "sedachain/pubkey/v1/query.proto",
}

func (m *QueryParamsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryParamsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryParamsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *QueryValidatorKeysRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryValidatorKeysRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryValidatorKeysRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ValidatorAddr) > 0 {
		i -= len(m.ValidatorAddr)
		copy(dAtA[i:], m.ValidatorAddr)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.ValidatorAddr)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryValidatorKeysResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryValidatorKeysResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryValidatorKeysResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.ValidatorPubKeys.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *QueryProvingSchemesRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryProvingSchemesRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryProvingSchemesRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryProvingSchemesResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryProvingSchemesResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryProvingSchemesResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ProvingSchemes) > 0 {
		for iNdEx := len(m.ProvingSchemes) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ProvingSchemes[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintQuery(dAtA []byte, offset int, v uint64) int {
	offset -= sovQuery(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *QueryParamsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryValidatorKeysRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ValidatorAddr)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryValidatorKeysResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.ValidatorPubKeys.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryProvingSchemesRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryProvingSchemesResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.ProvingSchemes) > 0 {
		for _, e := range m.ProvingSchemes {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryParamsRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryParamsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryParamsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryParamsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryValidatorKeysRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryValidatorKeysRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryValidatorKeysRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddr", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ValidatorAddr = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryValidatorKeysResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryValidatorKeysResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryValidatorKeysResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorPubKeys", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ValidatorPubKeys.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryProvingSchemesRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryProvingSchemesRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryProvingSchemesRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryProvingSchemesResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryProvingSchemesResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryProvingSchemesResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProvingSchemes", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ProvingSchemes = append(m.ProvingSchemes, ProvingScheme{})
			if err := m.ProvingSchemes[len(m.ProvingSchemes)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipQuery(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthQuery
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupQuery
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthQuery
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthQuery        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowQuery          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupQuery = fmt.Errorf("proto: unexpected end of group")
)
