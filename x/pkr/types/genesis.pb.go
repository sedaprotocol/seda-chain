// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sedachain/pkr/v1/genesis.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	types "github.com/cosmos/cosmos-sdk/codec/types"
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

// GenesisState defines pkr module's genesis state.
type GenesisState struct {
	ValidatorPubKeys []ValidatorPubKeys `protobuf:"bytes,1,rep,name=validator_pub_keys,json=validatorPubKeys,proto3" json:"validator_pub_keys"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_0a702f1e4b392dc9, []int{0}
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

func (m *GenesisState) GetValidatorPubKeys() []ValidatorPubKeys {
	if m != nil {
		return m.ValidatorPubKeys
	}
	return nil
}

// ValidatorPubKeys defines a validator's list of registered public keys
// primarily used in the x/pkr genesis state.
type ValidatorPubKeys struct {
	ValidatorAddr string            `protobuf:"bytes,1,opt,name=validator_addr,json=validatorAddr,proto3" json:"validator_addr,omitempty"`
	PubKeys       []IndexPubKeyPair `protobuf:"bytes,2,rep,name=pub_keys,json=pubKeys,proto3" json:"pub_keys"`
}

func (m *ValidatorPubKeys) Reset()         { *m = ValidatorPubKeys{} }
func (m *ValidatorPubKeys) String() string { return proto.CompactTextString(m) }
func (*ValidatorPubKeys) ProtoMessage()    {}
func (*ValidatorPubKeys) Descriptor() ([]byte, []int) {
	return fileDescriptor_0a702f1e4b392dc9, []int{1}
}
func (m *ValidatorPubKeys) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ValidatorPubKeys) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ValidatorPubKeys.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ValidatorPubKeys) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidatorPubKeys.Merge(m, src)
}
func (m *ValidatorPubKeys) XXX_Size() int {
	return m.Size()
}
func (m *ValidatorPubKeys) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidatorPubKeys.DiscardUnknown(m)
}

var xxx_messageInfo_ValidatorPubKeys proto.InternalMessageInfo

func (m *ValidatorPubKeys) GetValidatorAddr() string {
	if m != nil {
		return m.ValidatorAddr
	}
	return ""
}

func (m *ValidatorPubKeys) GetPubKeys() []IndexPubKeyPair {
	if m != nil {
		return m.PubKeys
	}
	return nil
}

// IndexPubKeyPair defines an index - public key pair.
type IndexPubKeyPair struct {
	Index  uint32     `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	PubKey *types.Any `protobuf:"bytes,2,opt,name=pub_key,json=pubKey,proto3" json:"pub_key,omitempty"`
}

func (m *IndexPubKeyPair) Reset()         { *m = IndexPubKeyPair{} }
func (m *IndexPubKeyPair) String() string { return proto.CompactTextString(m) }
func (*IndexPubKeyPair) ProtoMessage()    {}
func (*IndexPubKeyPair) Descriptor() ([]byte, []int) {
	return fileDescriptor_0a702f1e4b392dc9, []int{2}
}
func (m *IndexPubKeyPair) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *IndexPubKeyPair) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_IndexPubKeyPair.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *IndexPubKeyPair) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IndexPubKeyPair.Merge(m, src)
}
func (m *IndexPubKeyPair) XXX_Size() int {
	return m.Size()
}
func (m *IndexPubKeyPair) XXX_DiscardUnknown() {
	xxx_messageInfo_IndexPubKeyPair.DiscardUnknown(m)
}

var xxx_messageInfo_IndexPubKeyPair proto.InternalMessageInfo

func (m *IndexPubKeyPair) GetIndex() uint32 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *IndexPubKeyPair) GetPubKey() *types.Any {
	if m != nil {
		return m.PubKey
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "sedachain.pkr.v1.GenesisState")
	proto.RegisterType((*ValidatorPubKeys)(nil), "sedachain.pkr.v1.ValidatorPubKeys")
	proto.RegisterType((*IndexPubKeyPair)(nil), "sedachain.pkr.v1.IndexPubKeyPair")
}

func init() { proto.RegisterFile("sedachain/pkr/v1/genesis.proto", fileDescriptor_0a702f1e4b392dc9) }

var fileDescriptor_0a702f1e4b392dc9 = []byte{
	// 382 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x92, 0x31, 0xce, 0xda, 0x30,
	0x1c, 0xc5, 0x63, 0xda, 0x42, 0x6b, 0x4a, 0x8b, 0xa2, 0x0c, 0x29, 0x52, 0x53, 0x60, 0x62, 0xc1,
	0x11, 0xf4, 0x04, 0x64, 0x81, 0xaa, 0x0b, 0x0a, 0x12, 0x43, 0x17, 0xe4, 0x24, 0x26, 0x44, 0x40,
	0x6c, 0xd9, 0x4e, 0x44, 0x6e, 0xd1, 0x23, 0xf4, 0x10, 0x1c, 0x02, 0x75, 0x42, 0x9d, 0x3a, 0x55,
	0x15, 0x5c, 0xa4, 0xc2, 0x0e, 0xf4, 0x53, 0xbe, 0xcd, 0xef, 0xff, 0x9e, 0xfd, 0x7e, 0xb6, 0x0c,
	0x1d, 0x41, 0x22, 0x1c, 0x6e, 0x70, 0x92, 0xba, 0x6c, 0xcb, 0xdd, 0x7c, 0xe4, 0xc6, 0x24, 0x25,
	0x22, 0x11, 0x88, 0x71, 0x2a, 0xa9, 0xd9, 0x7e, 0xf8, 0x88, 0x6d, 0x39, 0xca, 0x47, 0x1d, 0x2b,
	0xa6, 0x31, 0x55, 0xa6, 0x7b, 0x5b, 0xe9, 0x5c, 0xe7, 0x43, 0x4c, 0x69, 0xbc, 0x23, 0xae, 0x52,
	0x41, 0xb6, 0x76, 0x71, 0x5a, 0xdc, 0xad, 0x90, 0x8a, 0x3d, 0x15, 0x2b, 0xbd, 0x47, 0x0b, 0x6d,
	0xf5, 0xd7, 0xf0, 0xed, 0x54, 0xd7, 0x2d, 0x24, 0x96, 0xc4, 0x5c, 0x42, 0x33, 0xc7, 0xbb, 0x24,
	0xc2, 0x92, 0xf2, 0x15, 0xcb, 0x82, 0xd5, 0x96, 0x14, 0xc2, 0x06, 0xdd, 0x17, 0x83, 0xe6, 0xb8,
	0x8f, 0xaa, 0x28, 0x68, 0x79, 0xcf, 0xce, 0xb3, 0xe0, 0x2b, 0x29, 0x84, 0xf7, 0xf2, 0xf4, 0xe7,
	0x93, 0xe1, 0xb7, 0xf3, 0xca, 0xbc, 0xff, 0x03, 0xc0, 0x76, 0x35, 0x6c, 0xce, 0xe0, 0xbb, 0xff,
	0x65, 0x38, 0x8a, 0xb8, 0x0d, 0xba, 0x60, 0xf0, 0xc6, 0xeb, 0xfd, 0x3a, 0x0e, 0x3f, 0x96, 0x98,
	0x8f, 0x4d, 0x93, 0x28, 0xe2, 0x44, 0x88, 0x85, 0xe4, 0x49, 0x1a, 0xfb, 0xad, 0xfc, 0xe9, 0xdc,
	0xf4, 0xe0, 0xeb, 0x07, 0x6c, 0x4d, 0xc1, 0xf6, 0x9e, 0xc3, 0x7e, 0x49, 0x23, 0x72, 0xd0, 0xdd,
	0x73, 0x9c, 0xf0, 0x92, 0xb5, 0xc1, 0x4a, 0x44, 0x06, 0xdf, 0x57, 0x12, 0xa6, 0x05, 0x5f, 0x25,
	0xb7, 0x91, 0xe2, 0x6a, 0xf9, 0x5a, 0x98, 0x53, 0xd8, 0x28, 0xcb, 0xec, 0x5a, 0x17, 0x0c, 0x9a,
	0x63, 0x0b, 0xe9, 0xb7, 0x47, 0xf7, 0xb7, 0x47, 0x93, 0xb4, 0xf0, 0xec, 0x9f, 0xc7, 0xa1, 0x55,
	0xde, 0x22, 0xe4, 0x05, 0x93, 0x14, 0xe9, 0xa3, 0xfd, 0xba, 0xae, 0xf4, 0x66, 0xa7, 0x8b, 0x03,
	0xce, 0x17, 0x07, 0xfc, 0xbd, 0x38, 0xe0, 0xfb, 0xd5, 0x31, 0xce, 0x57, 0xc7, 0xf8, 0x7d, 0x75,
	0x8c, 0x6f, 0x28, 0x4e, 0xe4, 0x26, 0x0b, 0x50, 0x48, 0xf7, 0xee, 0xed, 0x1e, 0xea, 0xe0, 0x90,
	0xee, 0x94, 0x18, 0xea, 0xdf, 0x72, 0x50, 0xff, 0x45, 0x16, 0x8c, 0x88, 0xa0, 0xae, 0x02, 0x9f,
	0xff, 0x05, 0x00, 0x00, 0xff, 0xff, 0xac, 0xa2, 0x9d, 0x9c, 0x4d, 0x02, 0x00, 0x00,
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
	if len(m.ValidatorPubKeys) > 0 {
		for iNdEx := len(m.ValidatorPubKeys) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ValidatorPubKeys[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *ValidatorPubKeys) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ValidatorPubKeys) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ValidatorPubKeys) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.PubKeys) > 0 {
		for iNdEx := len(m.PubKeys) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.PubKeys[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.ValidatorAddr) > 0 {
		i -= len(m.ValidatorAddr)
		copy(dAtA[i:], m.ValidatorAddr)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.ValidatorAddr)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *IndexPubKeyPair) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *IndexPubKeyPair) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *IndexPubKeyPair) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.PubKey != nil {
		{
			size, err := m.PubKey.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if m.Index != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Index))
		i--
		dAtA[i] = 0x8
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
	if len(m.ValidatorPubKeys) > 0 {
		for _, e := range m.ValidatorPubKeys {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *ValidatorPubKeys) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ValidatorAddr)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if len(m.PubKeys) > 0 {
		for _, e := range m.PubKeys {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *IndexPubKeyPair) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Index != 0 {
		n += 1 + sovGenesis(uint64(m.Index))
	}
	if m.PubKey != nil {
		l = m.PubKey.Size()
		n += 1 + l + sovGenesis(uint64(l))
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
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorPubKeys", wireType)
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
			m.ValidatorPubKeys = append(m.ValidatorPubKeys, ValidatorPubKeys{})
			if err := m.ValidatorPubKeys[len(m.ValidatorPubKeys)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *ValidatorPubKeys) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: ValidatorPubKeys: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ValidatorPubKeys: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddr", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ValidatorAddr = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PubKeys", wireType)
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
			m.PubKeys = append(m.PubKeys, IndexPubKeyPair{})
			if err := m.PubKeys[len(m.PubKeys)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *IndexPubKeyPair) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: IndexPubKeyPair: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: IndexPubKeyPair: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
			}
			m.Index = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return fmt.Errorf("proto: wrong wireType = %d for field PubKey", wireType)
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
			if m.PubKey == nil {
				m.PubKey = &types.Any{}
			}
			if err := m.PubKey.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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