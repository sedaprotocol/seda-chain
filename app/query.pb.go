// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sedachain/app/v1/query.proto

package app

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
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

// QuerySEDASignerStatusRequest is request type for the Query/SEDASignerStatus
// RPC method.
type QuerySEDASignerStatusRequest struct {
}

func (m *QuerySEDASignerStatusRequest) Reset()         { *m = QuerySEDASignerStatusRequest{} }
func (m *QuerySEDASignerStatusRequest) String() string { return proto.CompactTextString(m) }
func (*QuerySEDASignerStatusRequest) ProtoMessage()    {}
func (*QuerySEDASignerStatusRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_74921f2d23a2b089, []int{0}
}
func (m *QuerySEDASignerStatusRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QuerySEDASignerStatusRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QuerySEDASignerStatusRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QuerySEDASignerStatusRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QuerySEDASignerStatusRequest.Merge(m, src)
}
func (m *QuerySEDASignerStatusRequest) XXX_Size() int {
	return m.Size()
}
func (m *QuerySEDASignerStatusRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QuerySEDASignerStatusRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QuerySEDASignerStatusRequest proto.InternalMessageInfo

// QuerySEDASignerStatusResponse is response type for the Query/SEDASignerStatus
// RPC method.
type QuerySEDASignerStatusResponse struct {
	// ValidatorAddress is the address of the validator loaded in the signer.
	ValidatorAddress string `protobuf:"bytes,1,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	// SignerKeys is the list of keys loaded in the SEDA signer.
	SignerKeys []*SignerKey `protobuf:"bytes,2,rep,name=signer_keys,json=signerKeys,proto3" json:"signer_keys,omitempty"`
}

func (m *QuerySEDASignerStatusResponse) Reset()         { *m = QuerySEDASignerStatusResponse{} }
func (m *QuerySEDASignerStatusResponse) String() string { return proto.CompactTextString(m) }
func (*QuerySEDASignerStatusResponse) ProtoMessage()    {}
func (*QuerySEDASignerStatusResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_74921f2d23a2b089, []int{1}
}
func (m *QuerySEDASignerStatusResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QuerySEDASignerStatusResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QuerySEDASignerStatusResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QuerySEDASignerStatusResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QuerySEDASignerStatusResponse.Merge(m, src)
}
func (m *QuerySEDASignerStatusResponse) XXX_Size() int {
	return m.Size()
}
func (m *QuerySEDASignerStatusResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QuerySEDASignerStatusResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QuerySEDASignerStatusResponse proto.InternalMessageInfo

func (m *QuerySEDASignerStatusResponse) GetValidatorAddress() string {
	if m != nil {
		return m.ValidatorAddress
	}
	return ""
}

func (m *QuerySEDASignerStatusResponse) GetSignerKeys() []*SignerKey {
	if m != nil {
		return m.SignerKeys
	}
	return nil
}

// SignerKey is a key loaded in the SEDA signer.
type SignerKey struct {
	// Index is the index of the SEDA key.
	Index uint32 `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	// IndexName is the name of the SEDA key.
	IndexName string `protobuf:"bytes,2,opt,name=index_name,json=indexName,proto3" json:"index_name,omitempty"`
	// IsActive indicates whether the given SEDA key index is active.
	IsProvingSchemeActive bool `protobuf:"varint,3,opt,name=is_proving_scheme_active,json=isProvingSchemeActive,proto3" json:"is_proving_scheme_active,omitempty"`
	// PublicKey is the hex-encoded public key of the key loaded in
	// the SEDA signer.
	PublicKey string `protobuf:"bytes,4,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	// IsSynced indicates whether the keys loaded in the SEDA signer match
	// the keys registered in the pubkey module.
	IsSynced bool `protobuf:"varint,5,opt,name=is_synced,json=isSynced,proto3" json:"is_synced,omitempty"`
}

func (m *SignerKey) Reset()         { *m = SignerKey{} }
func (m *SignerKey) String() string { return proto.CompactTextString(m) }
func (*SignerKey) ProtoMessage()    {}
func (*SignerKey) Descriptor() ([]byte, []int) {
	return fileDescriptor_74921f2d23a2b089, []int{2}
}
func (m *SignerKey) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SignerKey) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SignerKey.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SignerKey) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignerKey.Merge(m, src)
}
func (m *SignerKey) XXX_Size() int {
	return m.Size()
}
func (m *SignerKey) XXX_DiscardUnknown() {
	xxx_messageInfo_SignerKey.DiscardUnknown(m)
}

var xxx_messageInfo_SignerKey proto.InternalMessageInfo

func (m *SignerKey) GetIndex() uint32 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *SignerKey) GetIndexName() string {
	if m != nil {
		return m.IndexName
	}
	return ""
}

func (m *SignerKey) GetIsProvingSchemeActive() bool {
	if m != nil {
		return m.IsProvingSchemeActive
	}
	return false
}

func (m *SignerKey) GetPublicKey() string {
	if m != nil {
		return m.PublicKey
	}
	return ""
}

func (m *SignerKey) GetIsSynced() bool {
	if m != nil {
		return m.IsSynced
	}
	return false
}

func init() {
	proto.RegisterType((*QuerySEDASignerStatusRequest)(nil), "sedachain.app.v1.QuerySEDASignerStatusRequest")
	proto.RegisterType((*QuerySEDASignerStatusResponse)(nil), "sedachain.app.v1.QuerySEDASignerStatusResponse")
	proto.RegisterType((*SignerKey)(nil), "sedachain.app.v1.SignerKey")
}

func init() { proto.RegisterFile("sedachain/app/v1/query.proto", fileDescriptor_74921f2d23a2b089) }

var fileDescriptor_74921f2d23a2b089 = []byte{
	// 478 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x52, 0x31, 0x6f, 0xd3, 0x40,
	0x14, 0xce, 0xa5, 0x04, 0xd5, 0x57, 0x21, 0x85, 0x53, 0x91, 0x4c, 0x9a, 0x58, 0x21, 0x43, 0x95,
	0x25, 0xb6, 0x5a, 0x86, 0x2e, 0x0c, 0xa4, 0x82, 0x09, 0xa9, 0x2a, 0xb6, 0xc4, 0xc0, 0x62, 0x9d,
	0xed, 0x93, 0x73, 0x6a, 0x7c, 0x77, 0xf5, 0x3b, 0x5b, 0x78, 0xe5, 0x17, 0x20, 0x31, 0x32, 0x33,
	0xb2, 0xf5, 0x47, 0x20, 0x16, 0x2a, 0x58, 0x18, 0x51, 0xc2, 0x0f, 0x41, 0xbe, 0x6b, 0x2b, 0x14,
	0x04, 0xea, 0x76, 0xef, 0x7d, 0xdf, 0xdd, 0xfb, 0xee, 0xfb, 0x1e, 0x1e, 0x02, 0xcb, 0x68, 0xba,
	0xa0, 0x5c, 0x04, 0x54, 0xa9, 0xa0, 0x3e, 0x08, 0xce, 0x2b, 0x56, 0x36, 0xbe, 0x2a, 0xa5, 0x96,
	0xa4, 0x7f, 0x83, 0xfa, 0x54, 0x29, 0xbf, 0x3e, 0x18, 0x0c, 0x73, 0x29, 0xf3, 0x25, 0x0b, 0xa8,
	0xe2, 0x01, 0x15, 0x42, 0x6a, 0xaa, 0xb9, 0x14, 0x60, 0xf9, 0x83, 0x87, 0xa9, 0x84, 0x42, 0x42,
	0x6c, 0xaa, 0xc0, 0x16, 0x16, 0x9a, 0x78, 0x78, 0xf8, 0xb2, 0x7d, 0x39, 0x7a, 0xfe, 0x6c, 0x1e,
	0xf1, 0x5c, 0xb0, 0x32, 0xd2, 0x54, 0x57, 0x10, 0xb2, 0xf3, 0x8a, 0x81, 0x9e, 0x7c, 0x42, 0x78,
	0xf4, 0x0f, 0x02, 0x28, 0x29, 0x80, 0x91, 0x13, 0x7c, 0xbf, 0xa6, 0x4b, 0x9e, 0x51, 0x2d, 0xcb,
	0x98, 0x66, 0x59, 0xc9, 0x00, 0x5c, 0x34, 0x46, 0x53, 0xe7, 0xf8, 0xd1, 0xb7, 0x8b, 0xd9, 0xe8,
	0x6a, 0xdc, 0xab, 0x6b, 0xce, 0xdc, 0x52, 0x22, 0x5d, 0x72, 0x91, 0x87, 0xfd, 0x7a, 0xa3, 0x4f,
	0x9e, 0xe0, 0x1d, 0x30, 0x73, 0xe2, 0x33, 0xd6, 0x80, 0xdb, 0x1d, 0x6f, 0x4d, 0x77, 0x0e, 0xf7,
	0xfc, 0xcd, 0x2f, 0xfb, 0x56, 0xcc, 0x0b, 0xd6, 0x84, 0x18, 0xae, 0x8f, 0x30, 0xf9, 0x8a, 0xb0,
	0x73, 0x83, 0x90, 0x5d, 0xdc, 0xe3, 0x22, 0x63, 0x6f, 0x8c, 0x9e, 0x7b, 0xa1, 0x2d, 0xc8, 0x08,
	0x63, 0x73, 0x88, 0x05, 0x2d, 0x98, 0xdb, 0x6d, 0xa5, 0x86, 0x8e, 0xe9, 0x9c, 0xd0, 0x82, 0x91,
	0x23, 0xec, 0x72, 0xe3, 0x55, 0xcd, 0x45, 0x1e, 0x43, 0xba, 0x60, 0x05, 0x8b, 0x69, 0xaa, 0x79,
	0xcd, 0xdc, 0xad, 0x31, 0x9a, 0x6e, 0x87, 0x0f, 0x38, 0x9c, 0x5a, 0x38, 0x32, 0xe8, 0xdc, 0x80,
	0xe4, 0x08, 0x63, 0x55, 0x25, 0x4b, 0x9e, 0xb6, 0xca, 0xdd, 0x3b, 0xc6, 0x02, 0xf7, 0xcb, 0xc5,
	0x6c, 0xf7, 0xca, 0x82, 0xb4, 0x6c, 0x94, 0x96, 0xfe, 0x69, 0x95, 0xb4, 0xaa, 0x1d, 0xcb, 0x6d,
	0x65, 0xee, 0x61, 0x87, 0x43, 0x0c, 0x8d, 0x48, 0x59, 0xe6, 0xf6, 0xcc, 0x88, 0x6d, 0x0e, 0x91,
	0xa9, 0x0f, 0x3f, 0x22, 0xdc, 0x33, 0x09, 0x90, 0x0f, 0x08, 0xf7, 0x37, 0x63, 0x20, 0xfe, 0xdf,
	0xce, 0xfc, 0x2f, 0xd0, 0x41, 0x70, 0x6b, 0xbe, 0xcd, 0x77, 0xb2, 0xff, 0xf6, 0xfb, 0xaf, 0xf7,
	0xdd, 0x31, 0xf1, 0x82, 0xf6, 0xe2, 0xcc, 0x2e, 0xa5, 0xaa, 0x92, 0x33, 0xd6, 0x98, 0x4e, 0x6c,
	0xdd, 0x3f, 0x7e, 0xfa, 0x79, 0xe5, 0xa1, 0xcb, 0x95, 0x87, 0x7e, 0xae, 0x3c, 0xf4, 0x6e, 0xed,
	0x75, 0x2e, 0xd7, 0x5e, 0xe7, 0xc7, 0xda, 0xeb, 0xbc, 0xde, 0xcf, 0xb9, 0x5e, 0x54, 0x89, 0x9f,
	0xca, 0xc2, 0xdc, 0x30, 0x9b, 0x97, 0xca, 0xe5, 0x9f, 0x0f, 0x52, 0xa5, 0x92, 0xbb, 0x06, 0x78,
	0xfc, 0x3b, 0x00, 0x00, 0xff, 0xff, 0xab, 0xc0, 0xca, 0xa4, 0xfd, 0x02, 0x00, 0x00,
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
	// SEDASignerStatus returns the status of the node's SEDA signer.
	SEDASignerStatus(ctx context.Context, in *QuerySEDASignerStatusRequest, opts ...grpc.CallOption) (*QuerySEDASignerStatusResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) SEDASignerStatus(ctx context.Context, in *QuerySEDASignerStatusRequest, opts ...grpc.CallOption) (*QuerySEDASignerStatusResponse, error) {
	out := new(QuerySEDASignerStatusResponse)
	err := c.cc.Invoke(ctx, "/sedachain.app.v1.Query/SEDASignerStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// SEDASignerStatus returns the status of the node's SEDA signer.
	SEDASignerStatus(context.Context, *QuerySEDASignerStatusRequest) (*QuerySEDASignerStatusResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) SEDASignerStatus(ctx context.Context, req *QuerySEDASignerStatusRequest) (*QuerySEDASignerStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SEDASignerStatus not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_SEDASignerStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QuerySEDASignerStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).SEDASignerStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sedachain.app.v1.Query/SEDASignerStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).SEDASignerStatus(ctx, req.(*QuerySEDASignerStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc
var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "sedachain.app.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SEDASignerStatus",
			Handler:    _Query_SEDASignerStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "sedachain/app/v1/query.proto",
}

func (m *QuerySEDASignerStatusRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QuerySEDASignerStatusRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QuerySEDASignerStatusRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QuerySEDASignerStatusResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QuerySEDASignerStatusResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QuerySEDASignerStatusResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.SignerKeys) > 0 {
		for iNdEx := len(m.SignerKeys) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.SignerKeys[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.ValidatorAddress) > 0 {
		i -= len(m.ValidatorAddress)
		copy(dAtA[i:], m.ValidatorAddress)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.ValidatorAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *SignerKey) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SignerKey) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SignerKey) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.IsSynced {
		i--
		if m.IsSynced {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x28
	}
	if len(m.PublicKey) > 0 {
		i -= len(m.PublicKey)
		copy(dAtA[i:], m.PublicKey)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.PublicKey)))
		i--
		dAtA[i] = 0x22
	}
	if m.IsProvingSchemeActive {
		i--
		if m.IsProvingSchemeActive {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x18
	}
	if len(m.IndexName) > 0 {
		i -= len(m.IndexName)
		copy(dAtA[i:], m.IndexName)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.IndexName)))
		i--
		dAtA[i] = 0x12
	}
	if m.Index != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.Index))
		i--
		dAtA[i] = 0x8
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
func (m *QuerySEDASignerStatusRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QuerySEDASignerStatusResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ValidatorAddress)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	if len(m.SignerKeys) > 0 {
		for _, e := range m.SignerKeys {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func (m *SignerKey) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Index != 0 {
		n += 1 + sovQuery(uint64(m.Index))
	}
	l = len(m.IndexName)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	if m.IsProvingSchemeActive {
		n += 2
	}
	l = len(m.PublicKey)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	if m.IsSynced {
		n += 2
	}
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QuerySEDASignerStatusRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QuerySEDASignerStatusRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QuerySEDASignerStatusRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QuerySEDASignerStatusResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QuerySEDASignerStatusResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QuerySEDASignerStatusResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddress", wireType)
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
			m.ValidatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SignerKeys", wireType)
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
			m.SignerKeys = append(m.SignerKeys, &SignerKey{})
			if err := m.SignerKeys[len(m.SignerKeys)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *SignerKey) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: SignerKey: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SignerKey: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
			}
			m.Index = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Index |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IndexName", wireType)
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
			m.IndexName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IsProvingSchemeActive", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.IsProvingSchemeActive = bool(v != 0)
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PublicKey", wireType)
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
			m.PublicKey = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IsSynced", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.IsSynced = bool(v != 0)
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
