// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sedachain/pubkey/v1/genesis.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
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

// GenesisState defines pubkey module's genesis state.
type GenesisState struct {
	Params           Params             `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	ValidatorPubKeys []ValidatorPubKeys `protobuf:"bytes,2,rep,name=validator_pub_keys,json=validatorPubKeys,proto3" json:"validator_pub_keys"`
	ProvingSchemes   []ProvingScheme    `protobuf:"bytes,3,rep,name=proving_schemes,json=provingSchemes,proto3" json:"proving_schemes"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_a68b70401eeae88a, []int{0}
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

func (m *GenesisState) GetValidatorPubKeys() []ValidatorPubKeys {
	if m != nil {
		return m.ValidatorPubKeys
	}
	return nil
}

func (m *GenesisState) GetProvingSchemes() []ProvingScheme {
	if m != nil {
		return m.ProvingSchemes
	}
	return nil
}

// ValidatorPubKeys defines a validator's list of registered public keys
// primarily used in the x/pubkey genesis state.
type ValidatorPubKeys struct {
	ValidatorAddr  string          `protobuf:"bytes,1,opt,name=validator_addr,json=validatorAddr,proto3" json:"validator_addr,omitempty"`
	IndexedPubKeys []IndexedPubKey `protobuf:"bytes,2,rep,name=indexed_pub_keys,json=indexedPubKeys,proto3" json:"indexed_pub_keys"`
}

func (m *ValidatorPubKeys) Reset()         { *m = ValidatorPubKeys{} }
func (m *ValidatorPubKeys) String() string { return proto.CompactTextString(m) }
func (*ValidatorPubKeys) ProtoMessage()    {}
func (*ValidatorPubKeys) Descriptor() ([]byte, []int) {
	return fileDescriptor_a68b70401eeae88a, []int{1}
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

func (m *ValidatorPubKeys) GetIndexedPubKeys() []IndexedPubKey {
	if m != nil {
		return m.IndexedPubKeys
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "sedachain.pubkey.v1.GenesisState")
	proto.RegisterType((*ValidatorPubKeys)(nil), "sedachain.pubkey.v1.ValidatorPubKeys")
}

func init() { proto.RegisterFile("sedachain/pubkey/v1/genesis.proto", fileDescriptor_a68b70401eeae88a) }

var fileDescriptor_a68b70401eeae88a = []byte{
	// 379 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x92, 0xc1, 0x4e, 0xea, 0x40,
	0x14, 0x86, 0xdb, 0xcb, 0x0d, 0xc9, 0x1d, 0xee, 0xe5, 0x92, 0xea, 0x02, 0x31, 0x56, 0x20, 0x31,
	0x61, 0x43, 0x1b, 0x70, 0xe5, 0x52, 0x36, 0x6a, 0xd8, 0x20, 0x24, 0x26, 0xba, 0x69, 0xa6, 0x9d,
	0x49, 0x99, 0x40, 0x3b, 0x4d, 0x4f, 0xdb, 0xd0, 0xb7, 0xf0, 0x51, 0x5c, 0xf8, 0x10, 0x2c, 0x89,
	0x2b, 0x57, 0xc6, 0xc0, 0x43, 0xb8, 0x35, 0xcc, 0xb4, 0x48, 0x48, 0x77, 0x9d, 0xff, 0x7c, 0xf3,
	0x9f, 0xbf, 0x67, 0x0e, 0x6a, 0x01, 0x25, 0xd8, 0x99, 0x62, 0xe6, 0x9b, 0x41, 0x6c, 0xcf, 0x68,
	0x6a, 0x26, 0x3d, 0xd3, 0xa5, 0x3e, 0x05, 0x06, 0x46, 0x10, 0xf2, 0x88, 0x6b, 0x47, 0x3b, 0xc4,
	0x90, 0x88, 0x91, 0xf4, 0x1a, 0xc7, 0x2e, 0x77, 0xb9, 0xa8, 0x9b, 0xdb, 0x2f, 0x89, 0x36, 0x4e,
	0x1c, 0x0e, 0x1e, 0x07, 0x4b, 0x16, 0xe4, 0x21, 0x2b, 0x35, 0x8b, 0x1a, 0x65, 0x7e, 0x82, 0x68,
	0x7f, 0xa9, 0xe8, 0xef, 0x8d, 0xec, 0x3c, 0x89, 0x70, 0x44, 0xb5, 0x2b, 0x54, 0x0e, 0x70, 0x88,
	0x3d, 0xa8, 0xab, 0x4d, 0xb5, 0x53, 0xe9, 0x9f, 0x1a, 0x05, 0x49, 0x8c, 0x91, 0x40, 0x06, 0xbf,
	0x97, 0x1f, 0xe7, 0xca, 0x38, 0xbb, 0xa0, 0x3d, 0x22, 0x2d, 0xc1, 0x73, 0x46, 0x70, 0xc4, 0x43,
	0x2b, 0x88, 0x6d, 0x6b, 0x46, 0x53, 0xa8, 0xff, 0x6a, 0x96, 0x3a, 0x95, 0xfe, 0x45, 0xa1, 0xcd,
	0x43, 0x8e, 0x8f, 0x62, 0x7b, 0x48, 0xd3, 0xdc, 0xb0, 0x96, 0x1c, 0xe8, 0xda, 0x3d, 0xfa, 0x1f,
	0x84, 0x3c, 0x61, 0xbe, 0x6b, 0x81, 0x33, 0xa5, 0x1e, 0x85, 0x7a, 0x49, 0xf8, 0xb6, 0x8b, 0xe3,
	0x49, 0x76, 0x22, 0xd0, 0xcc, 0xb4, 0x1a, 0xec, 0x8b, 0xd0, 0x7e, 0x51, 0x51, 0xed, 0xb0, 0xbf,
	0x76, 0x8b, 0xaa, 0x3f, 0xbf, 0x80, 0x09, 0x09, 0xc5, 0x14, 0xfe, 0x0c, 0x5a, 0x6f, 0xaf, 0xdd,
	0xb3, 0x6c, 0xb4, 0xbb, 0x4b, 0xd7, 0x84, 0x84, 0x14, 0x60, 0x12, 0x85, 0xcc, 0x77, 0xc7, 0xff,
	0x92, 0x7d, 0x5d, 0x1b, 0xa3, 0x1a, 0xf3, 0x09, 0x5d, 0x50, 0x72, 0x38, 0x8a, 0xe2, 0xc8, 0x77,
	0x12, 0x96, 0x41, 0xf2, 0xc8, 0x6c, 0x5f, 0x84, 0xc1, 0x70, 0xb9, 0xd6, 0xd5, 0xd5, 0x5a, 0x57,
	0x3f, 0xd7, 0xba, 0xfa, 0xbc, 0xd1, 0x95, 0xd5, 0x46, 0x57, 0xde, 0x37, 0xba, 0xf2, 0xd4, 0x73,
	0x59, 0x34, 0x8d, 0x6d, 0xc3, 0xe1, 0x9e, 0xb9, 0x75, 0x17, 0x8f, 0xeb, 0xf0, 0xb9, 0x38, 0x74,
	0xe5, 0x06, 0x2c, 0xf2, 0x1d, 0x88, 0xd2, 0x80, 0x82, 0x5d, 0x16, 0xcc, 0xe5, 0x77, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x6a, 0xb3, 0xd6, 0xe8, 0x8d, 0x02, 0x00, 0x00,
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
	if len(m.ProvingSchemes) > 0 {
		for iNdEx := len(m.ProvingSchemes) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ProvingSchemes[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.IndexedPubKeys) > 0 {
		for iNdEx := len(m.IndexedPubKeys) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.IndexedPubKeys[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.ValidatorPubKeys) > 0 {
		for _, e := range m.ValidatorPubKeys {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.ProvingSchemes) > 0 {
		for _, e := range m.ProvingSchemes {
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
	if len(m.IndexedPubKeys) > 0 {
		for _, e := range m.IndexedPubKeys {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
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
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProvingSchemes", wireType)
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
			m.ProvingSchemes = append(m.ProvingSchemes, ProvingScheme{})
			if err := m.ProvingSchemes[len(m.ProvingSchemes)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
				return fmt.Errorf("proto: wrong wireType = %d for field IndexedPubKeys", wireType)
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
			m.IndexedPubKeys = append(m.IndexedPubKeys, IndexedPubKey{})
			if err := m.IndexedPubKeys[len(m.IndexedPubKeys)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
