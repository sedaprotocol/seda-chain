// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sedachain/data_proxy/v1/genesis.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
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

// GenesisState defines data_proxy module's genesis state.
type GenesisState struct {
	Params           Params                 `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	DataProxyConfigs []DataProxyConfig      `protobuf:"bytes,2,rep,name=data_proxy_configs,json=dataProxyConfigs,proto3" json:"data_proxy_configs"`
	FeeUpdateQueue   []FeeUpdateQueueRecord `protobuf:"bytes,3,rep,name=fee_update_queue,json=feeUpdateQueue,proto3" json:"fee_update_queue"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_614b9aebcf526c4f, []int{0}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetDataProxyConfigs() []DataProxyConfig {
	if m != nil {
		return m.DataProxyConfigs
	}
	return nil
}

func (m *GenesisState) GetFeeUpdateQueue() []FeeUpdateQueueRecord {
	if m != nil {
		return m.FeeUpdateQueue
	}
	return nil
}

// DataProxyConfigs define the data proxy entries in the registry.
type DataProxyConfig struct {
	DataProxyPubkey []byte       `protobuf:"bytes,1,opt,name=data_proxy_pubkey,json=dataProxyPubkey,proto3" json:"data_proxy_pubkey,omitempty"`
	Config          *ProxyConfig `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
}

func (m *DataProxyConfig) Reset()         { *m = DataProxyConfig{} }
func (m *DataProxyConfig) String() string { return proto.CompactTextString(m) }
func (*DataProxyConfig) ProtoMessage()    {}
func (*DataProxyConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_614b9aebcf526c4f, []int{1}
}
func (m *DataProxyConfig) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *DataProxyConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_DataProxyConfig.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *DataProxyConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DataProxyConfig.Merge(m, src)
}
func (m *DataProxyConfig) XXX_Size() int {
	return m.Size()
}
func (m *DataProxyConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_DataProxyConfig.DiscardUnknown(m)
}

var xxx_messageInfo_DataProxyConfig proto.InternalMessageInfo

func (m *DataProxyConfig) GetDataProxyPubkey() []byte {
	if m != nil {
		return m.DataProxyPubkey
	}
	return nil
}

func (m *DataProxyConfig) GetConfig() *ProxyConfig {
	if m != nil {
		return m.Config
	}
	return nil
}

// FeeUpdateQueueRecord defines an entry in the data proxy update queue.
type FeeUpdateQueueRecord struct {
	DataProxyPubkey []byte `protobuf:"bytes,1,opt,name=data_proxy_pubkey,json=dataProxyPubkey,proto3" json:"data_proxy_pubkey,omitempty"`
	UpdateHeight    int64  `protobuf:"varint,2,opt,name=update_height,json=updateHeight,proto3" json:"update_height,omitempty"`
}

func (m *FeeUpdateQueueRecord) Reset()         { *m = FeeUpdateQueueRecord{} }
func (m *FeeUpdateQueueRecord) String() string { return proto.CompactTextString(m) }
func (*FeeUpdateQueueRecord) ProtoMessage()    {}
func (*FeeUpdateQueueRecord) Descriptor() ([]byte, []int) {
	return fileDescriptor_614b9aebcf526c4f, []int{2}
}
func (m *FeeUpdateQueueRecord) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FeeUpdateQueueRecord) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FeeUpdateQueueRecord.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FeeUpdateQueueRecord) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FeeUpdateQueueRecord.Merge(m, src)
}
func (m *FeeUpdateQueueRecord) XXX_Size() int {
	return m.Size()
}
func (m *FeeUpdateQueueRecord) XXX_DiscardUnknown() {
	xxx_messageInfo_FeeUpdateQueueRecord.DiscardUnknown(m)
}

var xxx_messageInfo_FeeUpdateQueueRecord proto.InternalMessageInfo

func (m *FeeUpdateQueueRecord) GetDataProxyPubkey() []byte {
	if m != nil {
		return m.DataProxyPubkey
	}
	return nil
}

func (m *FeeUpdateQueueRecord) GetUpdateHeight() int64 {
	if m != nil {
		return m.UpdateHeight
	}
	return 0
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "sedachain.data_proxy.v1.GenesisState")
	proto.RegisterType((*DataProxyConfig)(nil), "sedachain.data_proxy.v1.DataProxyConfig")
	proto.RegisterType((*FeeUpdateQueueRecord)(nil), "sedachain.data_proxy.v1.FeeUpdateQueueRecord")
}

func init() {
	proto.RegisterFile("sedachain/data_proxy/v1/genesis.proto", fileDescriptor_614b9aebcf526c4f)
}

var fileDescriptor_614b9aebcf526c4f = []byte{
	// 380 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0xcf, 0x4e, 0xea, 0x50,
	0x10, 0xc6, 0x5b, 0xb8, 0x61, 0x71, 0xe0, 0x5e, 0xb8, 0x0d, 0x89, 0x0d, 0x8b, 0x42, 0x50, 0x93,
	0xc6, 0x84, 0x36, 0x60, 0xdc, 0xe9, 0x06, 0x8d, 0xba, 0x13, 0x6b, 0xdc, 0x18, 0x4d, 0x73, 0x68,
	0x87, 0xb6, 0x51, 0x38, 0xb5, 0x3d, 0x25, 0x10, 0xdf, 0xc0, 0x95, 0x8f, 0xc5, 0x92, 0xa5, 0x2b,
	0x63, 0xe0, 0x45, 0x4c, 0xa7, 0x0d, 0x7f, 0x8c, 0x5d, 0xb8, 0xeb, 0x99, 0xfe, 0xe6, 0x9b, 0x6f,
	0x26, 0x1f, 0xd9, 0x0f, 0xc1, 0xa6, 0x96, 0x4b, 0xbd, 0x91, 0x6e, 0x53, 0x4e, 0x4d, 0x3f, 0x60,
	0x93, 0xa9, 0x3e, 0x6e, 0xeb, 0x0e, 0x8c, 0x20, 0xf4, 0x42, 0xcd, 0x0f, 0x18, 0x67, 0xd2, 0xce,
	0x0a, 0xd3, 0xd6, 0x98, 0x36, 0x6e, 0xd7, 0xaa, 0x0e, 0x73, 0x18, 0x32, 0x7a, 0xfc, 0x95, 0xe0,
	0x35, 0x35, 0x4b, 0x75, 0xa3, 0x19, 0xc9, 0xe6, 0x6b, 0x8e, 0x94, 0x2e, 0x92, 0x51, 0x37, 0x9c,
	0x72, 0x90, 0x4e, 0x48, 0xc1, 0xa7, 0x01, 0x1d, 0x86, 0xb2, 0xd8, 0x10, 0xd5, 0x62, 0xa7, 0xae,
	0x65, 0x8c, 0xd6, 0x7a, 0x88, 0x75, 0xff, 0xcc, 0x3e, 0xea, 0x82, 0x91, 0x36, 0x49, 0xf7, 0x44,
	0x5a, 0x53, 0xa6, 0xc5, 0x46, 0x03, 0xcf, 0x09, 0xe5, 0x5c, 0x23, 0xaf, 0x16, 0x3b, 0x6a, 0xa6,
	0xd4, 0x19, 0xe5, 0xb4, 0x17, 0x3f, 0x4e, 0xb1, 0x21, 0xd5, 0xac, 0xd8, 0xdb, 0xe5, 0x50, 0x7a,
	0x20, 0x95, 0x01, 0x80, 0x19, 0xf9, 0x36, 0xe5, 0x60, 0x3e, 0x47, 0x10, 0x81, 0x9c, 0x47, 0xed,
	0x56, 0xa6, 0xf6, 0x39, 0xc0, 0x2d, 0xf2, 0xd7, 0x31, 0x6e, 0x80, 0xc5, 0x02, 0x3b, 0x1d, 0xf0,
	0x6f, 0xb0, 0xf5, 0xaf, 0xf9, 0x42, 0xca, 0xdf, 0x9c, 0x48, 0x07, 0xe4, 0xff, 0xc6, 0x3e, 0x7e,
	0xd4, 0x7f, 0x84, 0x29, 0x5e, 0xa6, 0x64, 0x94, 0x57, 0xf6, 0x7a, 0x58, 0x96, 0x8e, 0x49, 0x21,
	0x59, 0x58, 0xce, 0xe1, 0xe9, 0xf6, 0xb2, 0x4f, 0xb7, 0x9e, 0x60, 0xa4, 0x3d, 0x4d, 0x87, 0x54,
	0x7f, 0xb2, 0xfa, 0x2b, 0x07, 0xbb, 0xe4, 0x6f, 0x7a, 0x1b, 0x17, 0x3c, 0xc7, 0xe5, 0x68, 0x24,
	0x6f, 0x94, 0x92, 0xe2, 0x25, 0xd6, 0xba, 0x57, 0xb3, 0x85, 0x22, 0xce, 0x17, 0x8a, 0xf8, 0xb9,
	0x50, 0xc4, 0xb7, 0xa5, 0x22, 0xcc, 0x97, 0x8a, 0xf0, 0xbe, 0x54, 0x84, 0xbb, 0x23, 0xc7, 0xe3,
	0x6e, 0xd4, 0xd7, 0x2c, 0x36, 0xd4, 0x63, 0xeb, 0x18, 0x11, 0x8b, 0x3d, 0xe1, 0xa3, 0x95, 0xe4,
	0x69, 0x82, 0x19, 0x6a, 0x25, 0x89, 0xe2, 0x53, 0x1f, 0xc2, 0x7e, 0x01, 0xb9, 0xc3, 0xaf, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x05, 0x85, 0x87, 0x38, 0xcc, 0x02, 0x00, 0x00,
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.FeeUpdateQueue) > 0 {
		for iNdEx := len(m.FeeUpdateQueue) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.FeeUpdateQueue[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.DataProxyConfigs) > 0 {
		for iNdEx := len(m.DataProxyConfigs) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DataProxyConfigs[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *DataProxyConfig) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DataProxyConfig) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *DataProxyConfig) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Config != nil {
		{
			size, err := m.Config.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if len(m.DataProxyPubkey) > 0 {
		i -= len(m.DataProxyPubkey)
		copy(dAtA[i:], m.DataProxyPubkey)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.DataProxyPubkey)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *FeeUpdateQueueRecord) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FeeUpdateQueueRecord) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *FeeUpdateQueueRecord) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.UpdateHeight != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.UpdateHeight))
		i--
		dAtA[i] = 0x10
	}
	if len(m.DataProxyPubkey) > 0 {
		i -= len(m.DataProxyPubkey)
		copy(dAtA[i:], m.DataProxyPubkey)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.DataProxyPubkey)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.DataProxyConfigs) > 0 {
		for _, e := range m.DataProxyConfigs {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.FeeUpdateQueue) > 0 {
		for _, e := range m.FeeUpdateQueue {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *DataProxyConfig) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.DataProxyPubkey)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.Config != nil {
		l = m.Config.Size()
		n += 1 + l + sovGenesis(uint64(l))
	}
	return n
}

func (m *FeeUpdateQueueRecord) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.DataProxyPubkey)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.UpdateHeight != 0 {
		n += 1 + sovGenesis(uint64(m.UpdateHeight))
	}
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataProxyConfigs", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DataProxyConfigs = append(m.DataProxyConfigs, DataProxyConfig{})
			if err := m.DataProxyConfigs[len(m.DataProxyConfigs)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FeeUpdateQueue", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.FeeUpdateQueue = append(m.FeeUpdateQueue, FeeUpdateQueueRecord{})
			if err := m.FeeUpdateQueue[len(m.FeeUpdateQueue)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *DataProxyConfig) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: DataProxyConfig: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DataProxyConfig: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataProxyPubkey", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DataProxyPubkey = append(m.DataProxyPubkey[:0], dAtA[iNdEx:postIndex]...)
			if m.DataProxyPubkey == nil {
				m.DataProxyPubkey = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Config", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Config == nil {
				m.Config = &ProxyConfig{}
			}
			if err := m.Config.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *FeeUpdateQueueRecord) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: FeeUpdateQueueRecord: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FeeUpdateQueueRecord: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataProxyPubkey", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DataProxyPubkey = append(m.DataProxyPubkey[:0], dAtA[iNdEx:postIndex]...)
			if m.DataProxyPubkey == nil {
				m.DataProxyPubkey = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UpdateHeight", wireType)
			}
			m.UpdateHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UpdateHeight |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis = fmt.Errorf("proto: unexpected end of group")
)
