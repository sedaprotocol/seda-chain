// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sedachain/vesting/v1/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/msgservice"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
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

// MsgCreateVestingAccount defines a message that creates a vesting account.
type MsgCreateVestingAccount struct {
	FromAddress string                                   `protobuf:"bytes,1,opt,name=from_address,json=fromAddress,proto3" json:"from_address,omitempty"`
	ToAddress   string                                   `protobuf:"bytes,2,opt,name=to_address,json=toAddress,proto3" json:"to_address,omitempty"`
	Amount      github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,3,rep,name=amount,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"amount"`
	// end of vesting as unix time (in seconds).
	EndTime int64 `protobuf:"varint,4,opt,name=end_time,json=endTime,proto3" json:"end_time,omitempty"`
}

func (m *MsgCreateVestingAccount) Reset()         { *m = MsgCreateVestingAccount{} }
func (m *MsgCreateVestingAccount) String() string { return proto.CompactTextString(m) }
func (*MsgCreateVestingAccount) ProtoMessage()    {}
func (*MsgCreateVestingAccount) Descriptor() ([]byte, []int) {
	return fileDescriptor_abaae49a55dd1e8c, []int{0}
}
func (m *MsgCreateVestingAccount) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCreateVestingAccount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCreateVestingAccount.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCreateVestingAccount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCreateVestingAccount.Merge(m, src)
}
func (m *MsgCreateVestingAccount) XXX_Size() int {
	return m.Size()
}
func (m *MsgCreateVestingAccount) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCreateVestingAccount.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCreateVestingAccount proto.InternalMessageInfo

func (m *MsgCreateVestingAccount) GetFromAddress() string {
	if m != nil {
		return m.FromAddress
	}
	return ""
}

func (m *MsgCreateVestingAccount) GetToAddress() string {
	if m != nil {
		return m.ToAddress
	}
	return ""
}

func (m *MsgCreateVestingAccount) GetAmount() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Amount
	}
	return nil
}

func (m *MsgCreateVestingAccount) GetEndTime() int64 {
	if m != nil {
		return m.EndTime
	}
	return 0
}

// MsgCreateVestingAccountResponse defines the CreateVestingAccount response type.
type MsgCreateVestingAccountResponse struct {
}

func (m *MsgCreateVestingAccountResponse) Reset()         { *m = MsgCreateVestingAccountResponse{} }
func (m *MsgCreateVestingAccountResponse) String() string { return proto.CompactTextString(m) }
func (*MsgCreateVestingAccountResponse) ProtoMessage()    {}
func (*MsgCreateVestingAccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_abaae49a55dd1e8c, []int{1}
}
func (m *MsgCreateVestingAccountResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCreateVestingAccountResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCreateVestingAccountResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCreateVestingAccountResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCreateVestingAccountResponse.Merge(m, src)
}
func (m *MsgCreateVestingAccountResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgCreateVestingAccountResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCreateVestingAccountResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCreateVestingAccountResponse proto.InternalMessageInfo

// MsgClawback defines a message that returns the vesting funds to the funder.
type MsgClawback struct {
	// funder_address is the address which funded the account.
	FunderAddress string `protobuf:"bytes,1,opt,name=funder_address,json=funderAddress,proto3" json:"funder_address,omitempty"`
	// account_address is the address of the vesting to claw back from.
	AccountAddress string `protobuf:"bytes,2,opt,name=account_address,json=accountAddress,proto3" json:"account_address,omitempty"`
}

func (m *MsgClawback) Reset()         { *m = MsgClawback{} }
func (m *MsgClawback) String() string { return proto.CompactTextString(m) }
func (*MsgClawback) ProtoMessage()    {}
func (*MsgClawback) Descriptor() ([]byte, []int) {
	return fileDescriptor_abaae49a55dd1e8c, []int{2}
}
func (m *MsgClawback) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgClawback) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgClawback.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgClawback) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgClawback.Merge(m, src)
}
func (m *MsgClawback) XXX_Size() int {
	return m.Size()
}
func (m *MsgClawback) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgClawback.DiscardUnknown(m)
}

var xxx_messageInfo_MsgClawback proto.InternalMessageInfo

func (m *MsgClawback) GetFunderAddress() string {
	if m != nil {
		return m.FunderAddress
	}
	return ""
}

func (m *MsgClawback) GetAccountAddress() string {
	if m != nil {
		return m.AccountAddress
	}
	return ""
}

// MsgClawbackResponse defines the MsgClawback response type.
type MsgClawbackResponse struct {
	ClawedUnbonded  github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,1,rep,name=clawed_unbonded,json=clawedUnbonded,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"clawed_unbonded"`
	ClawedUnbonding github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,2,rep,name=clawed_unbonding,json=clawedUnbonding,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"clawed_unbonding"`
	ClawedBonded    github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,3,rep,name=clawed_bonded,json=clawedBonded,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"clawed_bonded"`
}

func (m *MsgClawbackResponse) Reset()         { *m = MsgClawbackResponse{} }
func (m *MsgClawbackResponse) String() string { return proto.CompactTextString(m) }
func (*MsgClawbackResponse) ProtoMessage()    {}
func (*MsgClawbackResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_abaae49a55dd1e8c, []int{3}
}
func (m *MsgClawbackResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgClawbackResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgClawbackResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgClawbackResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgClawbackResponse.Merge(m, src)
}
func (m *MsgClawbackResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgClawbackResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgClawbackResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgClawbackResponse proto.InternalMessageInfo

func (m *MsgClawbackResponse) GetClawedUnbonded() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.ClawedUnbonded
	}
	return nil
}

func (m *MsgClawbackResponse) GetClawedUnbonding() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.ClawedUnbonding
	}
	return nil
}

func (m *MsgClawbackResponse) GetClawedBonded() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.ClawedBonded
	}
	return nil
}

func init() {
	proto.RegisterType((*MsgCreateVestingAccount)(nil), "sedachain.vesting.v1.MsgCreateVestingAccount")
	proto.RegisterType((*MsgCreateVestingAccountResponse)(nil), "sedachain.vesting.v1.MsgCreateVestingAccountResponse")
	proto.RegisterType((*MsgClawback)(nil), "sedachain.vesting.v1.MsgClawback")
	proto.RegisterType((*MsgClawbackResponse)(nil), "sedachain.vesting.v1.MsgClawbackResponse")
}

func init() { proto.RegisterFile("sedachain/vesting/v1/tx.proto", fileDescriptor_abaae49a55dd1e8c) }

var fileDescriptor_abaae49a55dd1e8c = []byte{
	// 611 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x54, 0xbf, 0x6f, 0xd3, 0x40,
	0x18, 0x8d, 0x13, 0x28, 0xed, 0x25, 0x4d, 0xc0, 0x8d, 0x54, 0x27, 0x12, 0x4e, 0x1a, 0x81, 0x08,
	0x91, 0x62, 0x93, 0x20, 0x84, 0x54, 0x58, 0x9a, 0x4a, 0x4c, 0x74, 0x31, 0x3f, 0x84, 0x58, 0xa2,
	0xb3, 0x7d, 0x75, 0x4f, 0x8d, 0xef, 0x22, 0xdf, 0x25, 0x6d, 0x24, 0x06, 0xc4, 0xc8, 0xc4, 0xcc,
	0xc4, 0x88, 0x98, 0x32, 0xf0, 0x47, 0x74, 0xac, 0x98, 0x98, 0x00, 0x25, 0x48, 0xe1, 0x1f, 0x60,
	0x47, 0xf6, 0x9d, 0x43, 0x52, 0x35, 0x8a, 0x2a, 0x75, 0x49, 0xec, 0xef, 0xbd, 0xf7, 0xbd, 0xcf,
	0xef, 0xec, 0x0f, 0xdc, 0x64, 0xc8, 0x85, 0xce, 0x01, 0xc4, 0xc4, 0xec, 0x23, 0xc6, 0x31, 0xf1,
	0xcc, 0x7e, 0xc3, 0xe4, 0xc7, 0x46, 0x37, 0xa0, 0x9c, 0xaa, 0xf9, 0x29, 0x6c, 0x48, 0xd8, 0xe8,
	0x37, 0x8a, 0x79, 0x8f, 0x7a, 0x34, 0x22, 0x98, 0xe1, 0x95, 0xe0, 0x16, 0x75, 0x87, 0x32, 0x9f,
	0x32, 0xd3, 0x86, 0x0c, 0x99, 0xfd, 0x86, 0x8d, 0x38, 0x6c, 0x98, 0x0e, 0xc5, 0x44, 0xe2, 0x05,
	0x81, 0xb7, 0x85, 0x50, 0xdc, 0x48, 0xe8, 0x96, 0x94, 0xfe, 0x1f, 0x41, 0xa8, 0x63, 0x4f, 0xc1,
	0xda, 0x94, 0x2c, 0x9f, 0x45, 0x43, 0xfa, 0x2c, 0x06, 0x6e, 0x40, 0x1f, 0x13, 0x6a, 0x46, 0xbf,
	0xa2, 0x54, 0x19, 0x27, 0xc1, 0xe6, 0x1e, 0xf3, 0x76, 0x03, 0x04, 0x39, 0x7a, 0x29, 0xda, 0xec,
	0x38, 0x0e, 0xed, 0x11, 0xae, 0x3e, 0x02, 0x99, 0xfd, 0x80, 0xfa, 0x6d, 0xe8, 0xba, 0x01, 0x62,
	0x4c, 0x53, 0xca, 0x4a, 0x75, 0xad, 0xa5, 0x7d, 0xfb, 0x5a, 0xcf, 0xcb, 0xa9, 0x76, 0x04, 0xf2,
	0x8c, 0x07, 0x98, 0x78, 0x56, 0x3a, 0x64, 0xcb, 0x92, 0xfa, 0x10, 0x00, 0x4e, 0xa7, 0xd2, 0xe4,
	0x12, 0xe9, 0x1a, 0xa7, 0xb1, 0x70, 0x00, 0x56, 0xa0, 0x1f, 0xfa, 0x6b, 0xa9, 0x72, 0xaa, 0x9a,
	0x6e, 0x16, 0x0c, 0xa9, 0x08, 0xf3, 0x32, 0xe4, 0x13, 0x1b, 0xbb, 0x14, 0x93, 0xd6, 0x93, 0x93,
	0x1f, 0xa5, 0xc4, 0x97, 0x9f, 0xa5, 0xaa, 0x87, 0xf9, 0x41, 0xcf, 0x36, 0x1c, 0xea, 0xcb, 0xbc,
	0xe4, 0x5f, 0x9d, 0xb9, 0x87, 0x26, 0x1f, 0x74, 0x11, 0x8b, 0x04, 0xec, 0xe3, 0x64, 0x58, 0xcb,
	0x74, 0x90, 0x07, 0x9d, 0x41, 0x3b, 0x4c, 0x9c, 0x7d, 0x9e, 0x0c, 0x6b, 0x8a, 0x25, 0x0d, 0xd5,
	0x02, 0x58, 0x45, 0xc4, 0x6d, 0x73, 0xec, 0x23, 0xed, 0x4a, 0x59, 0xa9, 0xa6, 0xac, 0x6b, 0x88,
	0xb8, 0xcf, 0xb1, 0x8f, 0xb6, 0x1f, 0xff, 0xf9, 0x54, 0x52, 0xde, 0x85, 0xf2, 0xd9, 0x48, 0xde,
	0x4f, 0x86, 0xb5, 0xca, 0x8c, 0xd5, 0x82, 0x24, 0x2b, 0x5b, 0xa0, 0xb4, 0x00, 0xb2, 0x10, 0xeb,
	0x52, 0xc2, 0x50, 0x25, 0x00, 0xe9, 0x90, 0xd2, 0x81, 0x47, 0x36, 0x74, 0x0e, 0xd5, 0xdb, 0x20,
	0xbb, 0xdf, 0x23, 0x2e, 0x0a, 0xe6, 0xd3, 0xb7, 0xd6, 0x45, 0x35, 0x0e, 0xeb, 0x0e, 0xc8, 0x41,
	0xd1, 0x68, 0x3e, 0x6a, 0x2b, 0x2b, 0xcb, 0x92, 0xb8, 0xbd, 0x11, 0xce, 0x7e, 0xa6, 0x65, 0xe5,
	0x6f, 0x12, 0x6c, 0xcc, 0x98, 0xc6, 0xb3, 0xa8, 0x1c, 0xe4, 0x9c, 0x0e, 0x3c, 0x42, 0x6e, 0xbb,
	0x47, 0x6c, 0x4a, 0x5c, 0xe4, 0x6a, 0xca, 0xb2, 0xb3, 0xb8, 0x77, 0xd1, 0xb3, 0xb0, 0xb2, 0xc2,
	0xe3, 0x85, 0xb4, 0x50, 0xfb, 0xe0, 0xfa, 0x9c, 0x2b, 0x26, 0x9e, 0x96, 0xbc, 0x7c, 0xdb, 0xdc,
	0xac, 0x2d, 0x26, 0x9e, 0xda, 0x05, 0xeb, 0xd2, 0x57, 0x3e, 0x6b, 0xea, 0xf2, 0x4d, 0x33, 0xc2,
	0xa1, 0x15, 0x19, 0x34, 0x7f, 0x2b, 0x20, 0xb5, 0xc7, 0x3c, 0xf5, 0x0d, 0xc8, 0x9f, 0xfb, 0xe1,
	0xd5, 0x8d, 0xf3, 0xd6, 0x89, 0xb1, 0xe0, 0x15, 0x2a, 0x3e, 0xb8, 0x10, 0x7d, 0x7a, 0xca, 0xaf,
	0xc0, 0xea, 0xf4, 0x75, 0xdb, 0x5a, 0xdc, 0x42, 0x52, 0x8a, 0x77, 0x97, 0x52, 0xe2, 0xce, 0xc5,
	0xab, 0x6f, 0xc3, 0xcf, 0xaa, 0xf5, 0xf4, 0x64, 0xa4, 0x2b, 0xa7, 0x23, 0x5d, 0xf9, 0x35, 0xd2,
	0x95, 0x0f, 0x63, 0x3d, 0x71, 0x3a, 0xd6, 0x13, 0xdf, 0xc7, 0x7a, 0xe2, 0x75, 0x73, 0x26, 0xb8,
	0xb0, 0x6b, 0xb4, 0x8b, 0x1c, 0xda, 0x89, 0x6e, 0xea, 0x62, 0xcd, 0x1e, 0x4f, 0xb7, 0x5c, 0x14,
	0xa4, 0xbd, 0x12, 0x91, 0xee, 0xff, 0x0b, 0x00, 0x00, 0xff, 0xff, 0x24, 0x17, 0x9e, 0x14, 0x8a,
	0x05, 0x00, 0x00,
}

func (this *MsgCreateVestingAccount) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*MsgCreateVestingAccount)
	if !ok {
		that2, ok := that.(MsgCreateVestingAccount)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.FromAddress != that1.FromAddress {
		return false
	}
	if this.ToAddress != that1.ToAddress {
		return false
	}
	if len(this.Amount) != len(that1.Amount) {
		return false
	}
	for i := range this.Amount {
		if !this.Amount[i].Equal(&that1.Amount[i]) {
			return false
		}
	}
	if this.EndTime != that1.EndTime {
		return false
	}
	return true
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	// CreateVestingAccount creates a new vesting account.
	CreateVestingAccount(ctx context.Context, in *MsgCreateVestingAccount, opts ...grpc.CallOption) (*MsgCreateVestingAccountResponse, error)
	// Clawback returns the vesting funds back to the funder.
	Clawback(ctx context.Context, in *MsgClawback, opts ...grpc.CallOption) (*MsgClawbackResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) CreateVestingAccount(ctx context.Context, in *MsgCreateVestingAccount, opts ...grpc.CallOption) (*MsgCreateVestingAccountResponse, error) {
	out := new(MsgCreateVestingAccountResponse)
	err := c.cc.Invoke(ctx, "/sedachain.vesting.v1.Msg/CreateVestingAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) Clawback(ctx context.Context, in *MsgClawback, opts ...grpc.CallOption) (*MsgClawbackResponse, error) {
	out := new(MsgClawbackResponse)
	err := c.cc.Invoke(ctx, "/sedachain.vesting.v1.Msg/Clawback", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// CreateVestingAccount creates a new vesting account.
	CreateVestingAccount(context.Context, *MsgCreateVestingAccount) (*MsgCreateVestingAccountResponse, error)
	// Clawback returns the vesting funds back to the funder.
	Clawback(context.Context, *MsgClawback) (*MsgClawbackResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) CreateVestingAccount(ctx context.Context, req *MsgCreateVestingAccount) (*MsgCreateVestingAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateVestingAccount not implemented")
}
func (*UnimplementedMsgServer) Clawback(ctx context.Context, req *MsgClawback) (*MsgClawbackResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Clawback not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_CreateVestingAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateVestingAccount)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateVestingAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sedachain.vesting.v1.Msg/CreateVestingAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateVestingAccount(ctx, req.(*MsgCreateVestingAccount))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_Clawback_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgClawback)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Clawback(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sedachain.vesting.v1.Msg/Clawback",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Clawback(ctx, req.(*MsgClawback))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "sedachain.vesting.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateVestingAccount",
			Handler:    _Msg_CreateVestingAccount_Handler,
		},
		{
			MethodName: "Clawback",
			Handler:    _Msg_Clawback_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "sedachain/vesting/v1/tx.proto",
}

func (m *MsgCreateVestingAccount) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCreateVestingAccount) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCreateVestingAccount) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.EndTime != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.EndTime))
		i--
		dAtA[i] = 0x20
	}
	if len(m.Amount) > 0 {
		for iNdEx := len(m.Amount) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Amount[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTx(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.ToAddress) > 0 {
		i -= len(m.ToAddress)
		copy(dAtA[i:], m.ToAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ToAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.FromAddress) > 0 {
		i -= len(m.FromAddress)
		copy(dAtA[i:], m.FromAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.FromAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgCreateVestingAccountResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCreateVestingAccountResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCreateVestingAccountResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *MsgClawback) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgClawback) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgClawback) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.AccountAddress) > 0 {
		i -= len(m.AccountAddress)
		copy(dAtA[i:], m.AccountAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.AccountAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.FunderAddress) > 0 {
		i -= len(m.FunderAddress)
		copy(dAtA[i:], m.FunderAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.FunderAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgClawbackResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgClawbackResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgClawbackResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ClawedBonded) > 0 {
		for iNdEx := len(m.ClawedBonded) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ClawedBonded[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTx(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.ClawedUnbonding) > 0 {
		for iNdEx := len(m.ClawedUnbonding) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ClawedUnbonding[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTx(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.ClawedUnbonded) > 0 {
		for iNdEx := len(m.ClawedUnbonded) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ClawedUnbonded[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTx(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgCreateVestingAccount) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.FromAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ToAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if len(m.Amount) > 0 {
		for _, e := range m.Amount {
			l = e.Size()
			n += 1 + l + sovTx(uint64(l))
		}
	}
	if m.EndTime != 0 {
		n += 1 + sovTx(uint64(m.EndTime))
	}
	return n
}

func (m *MsgCreateVestingAccountResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *MsgClawback) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.FunderAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.AccountAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgClawbackResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.ClawedUnbonded) > 0 {
		for _, e := range m.ClawedUnbonded {
			l = e.Size()
			n += 1 + l + sovTx(uint64(l))
		}
	}
	if len(m.ClawedUnbonding) > 0 {
		for _, e := range m.ClawedUnbonding {
			l = e.Size()
			n += 1 + l + sovTx(uint64(l))
		}
	}
	if len(m.ClawedBonded) > 0 {
		for _, e := range m.ClawedBonded {
			l = e.Size()
			n += 1 + l + sovTx(uint64(l))
		}
	}
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgCreateVestingAccount) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgCreateVestingAccount: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCreateVestingAccount: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FromAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.FromAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ToAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ToAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Amount = append(m.Amount, types.Coin{})
			if err := m.Amount[len(m.Amount)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field EndTime", wireType)
			}
			m.EndTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.EndTime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgCreateVestingAccountResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgCreateVestingAccountResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCreateVestingAccountResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgClawback) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgClawback: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgClawback: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FunderAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.FunderAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AccountAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AccountAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgClawbackResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgClawbackResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgClawbackResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClawedUnbonded", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ClawedUnbonded = append(m.ClawedUnbonded, types.Coin{})
			if err := m.ClawedUnbonded[len(m.ClawedUnbonded)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClawedUnbonding", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ClawedUnbonding = append(m.ClawedUnbonding, types.Coin{})
			if err := m.ClawedUnbonding[len(m.ClawedUnbonding)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClawedBonded", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ClawedBonded = append(m.ClawedBonded, types.Coin{})
			if err := m.ClawedBonded[len(m.ClawedBonded)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
