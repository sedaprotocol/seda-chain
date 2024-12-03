// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sedachain/pubkey/v1/pubkey.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
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

// IndexPubKeyPair defines an index - public key pair.
type IndexedPubKey struct {
	Index  uint32 `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	PubKey []byte `protobuf:"bytes,2,opt,name=pub_key,json=pubKey,proto3" json:"pub_key,omitempty"`
}

func (m *IndexedPubKey) Reset()         { *m = IndexedPubKey{} }
func (m *IndexedPubKey) String() string { return proto.CompactTextString(m) }
func (*IndexedPubKey) ProtoMessage()    {}
func (*IndexedPubKey) Descriptor() ([]byte, []int) {
	return fileDescriptor_a51ebcd05a6c14e0, []int{0}
}
func (m *IndexedPubKey) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *IndexedPubKey) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_IndexedPubKey.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *IndexedPubKey) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IndexedPubKey.Merge(m, src)
}
func (m *IndexedPubKey) XXX_Size() int {
	return m.Size()
}
func (m *IndexedPubKey) XXX_DiscardUnknown() {
	xxx_messageInfo_IndexedPubKey.DiscardUnknown(m)
}

var xxx_messageInfo_IndexedPubKey proto.InternalMessageInfo

func (m *IndexedPubKey) GetIndex() uint32 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *IndexedPubKey) GetPubKey() []byte {
	if m != nil {
		return m.PubKey
	}
	return nil
}

// ProvingScheme defines a proving scheme.
type ProvingScheme struct {
	// index is the SEDA key index.
	Index uint32 `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	// is_activated indicates if the proving scheme has been activated.
	IsActivated bool `protobuf:"varint,2,opt,name=is_activated,json=isActivated,proto3" json:"is_activated,omitempty"`
	// activation_height is the height at which the proving scheme is to
	// be activated. This field is set to -1 by default until the public
	// key registration rate reaches the activation threshold and is reset
	// if the public key registration rate goes below the threshold before
	// the scheme is activated.
	ActivationHeight int64 `protobuf:"varint,3,opt,name=activation_height,json=activationHeight,proto3" json:"activation_height,omitempty"`
}

func (m *ProvingScheme) Reset()         { *m = ProvingScheme{} }
func (m *ProvingScheme) String() string { return proto.CompactTextString(m) }
func (*ProvingScheme) ProtoMessage()    {}
func (*ProvingScheme) Descriptor() ([]byte, []int) {
	return fileDescriptor_a51ebcd05a6c14e0, []int{1}
}
func (m *ProvingScheme) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ProvingScheme) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ProvingScheme.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ProvingScheme) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProvingScheme.Merge(m, src)
}
func (m *ProvingScheme) XXX_Size() int {
	return m.Size()
}
func (m *ProvingScheme) XXX_DiscardUnknown() {
	xxx_messageInfo_ProvingScheme.DiscardUnknown(m)
}

var xxx_messageInfo_ProvingScheme proto.InternalMessageInfo

func (m *ProvingScheme) GetIndex() uint32 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *ProvingScheme) GetIsActivated() bool {
	if m != nil {
		return m.IsActivated
	}
	return false
}

func (m *ProvingScheme) GetActivationHeight() int64 {
	if m != nil {
		return m.ActivationHeight
	}
	return 0
}

// Params defines the parameters for the pubkey module.
type Params struct {
	// activation_lag is the number of blocks to wait before activating a
	// proving scheme.
	ActivationLag int64 `protobuf:"varint,1,opt,name=activation_lag,json=activationLag,proto3" json:"activation_lag,omitempty"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_a51ebcd05a6c14e0, []int{2}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetActivationLag() int64 {
	if m != nil {
		return m.ActivationLag
	}
	return 0
}

func init() {
	proto.RegisterType((*IndexedPubKey)(nil), "sedachain.pubkey.v1.IndexedPubKey")
	proto.RegisterType((*ProvingScheme)(nil), "sedachain.pubkey.v1.ProvingScheme")
	proto.RegisterType((*Params)(nil), "sedachain.pubkey.v1.Params")
}

func init() { proto.RegisterFile("sedachain/pubkey/v1/pubkey.proto", fileDescriptor_a51ebcd05a6c14e0) }

var fileDescriptor_a51ebcd05a6c14e0 = []byte{
	// 315 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x91, 0xc1, 0x4a, 0xc3, 0x30,
	0x1c, 0xc6, 0x17, 0x87, 0x53, 0xe2, 0x2a, 0x5a, 0x77, 0xa8, 0x1e, 0x42, 0x1d, 0x08, 0x03, 0x59,
	0x4b, 0xf1, 0x09, 0xdc, 0x49, 0x99, 0x87, 0x51, 0x2f, 0xe2, 0xa5, 0xa4, 0x69, 0x68, 0xc3, 0xd6,
	0x26, 0x34, 0x69, 0x59, 0xdf, 0xc2, 0x87, 0xf1, 0x21, 0xc4, 0xd3, 0x8e, 0x1e, 0xa5, 0x7d, 0x11,
	0x59, 0xda, 0x39, 0x2f, 0x9e, 0x92, 0xef, 0xf7, 0xff, 0xf2, 0x05, 0xfe, 0x1f, 0xb4, 0x25, 0x8d,
	0x30, 0x49, 0x30, 0xcb, 0x5c, 0x51, 0x84, 0x4b, 0x5a, 0xb9, 0xa5, 0xd7, 0xdd, 0x1c, 0x91, 0x73,
	0xc5, 0xcd, 0x8b, 0x5f, 0x87, 0xd3, 0xf1, 0xd2, 0xbb, 0xba, 0x24, 0x5c, 0xa6, 0x5c, 0x06, 0xda,
	0xe2, 0xb6, 0xa2, 0xf5, 0x8f, 0x5f, 0xa0, 0xf1, 0x98, 0x45, 0x74, 0x4d, 0xa3, 0x45, 0x11, 0xce,
	0x69, 0x65, 0x8e, 0xe0, 0x21, 0xdb, 0x02, 0x0b, 0xd8, 0x60, 0x62, 0xf8, 0xad, 0x30, 0x3d, 0x78,
	0x24, 0x8a, 0x30, 0x58, 0xd2, 0xca, 0x3a, 0xb0, 0xc1, 0x64, 0x38, 0xb3, 0x3e, 0xdf, 0xa7, 0xa3,
	0x2e, 0x89, 0xe4, 0x95, 0x50, 0xdc, 0x69, 0x03, 0xfc, 0x81, 0xd0, 0xe7, 0xb8, 0x80, 0xc6, 0x22,
	0xe7, 0x25, 0xcb, 0xe2, 0x67, 0x92, 0xd0, 0x94, 0xfe, 0x93, 0x7c, 0x0d, 0x87, 0x4c, 0x06, 0x98,
	0x28, 0x56, 0x62, 0x45, 0x23, 0x1d, 0x7f, 0xec, 0x9f, 0x30, 0x79, 0xbf, 0x43, 0xe6, 0x2d, 0x3c,
	0xef, 0xe6, 0x8c, 0x67, 0x41, 0x42, 0x59, 0x9c, 0x28, 0xab, 0x6f, 0x83, 0x49, 0xdf, 0x3f, 0xdb,
	0x0f, 0x1e, 0x34, 0x1f, 0xbb, 0x70, 0xb0, 0xc0, 0x39, 0x4e, 0xa5, 0x79, 0x03, 0x4f, 0xff, 0x3c,
	0x5b, 0xe1, 0x58, 0x7f, 0xdc, 0xf7, 0x8d, 0x3d, 0x7d, 0xc2, 0xf1, 0x6c, 0xfe, 0x51, 0x23, 0xb0,
	0xa9, 0x11, 0xf8, 0xae, 0x11, 0x78, 0x6b, 0x50, 0x6f, 0xd3, 0xa0, 0xde, 0x57, 0x83, 0x7a, 0xaf,
	0x5e, 0xcc, 0x54, 0x52, 0x84, 0x0e, 0xe1, 0xa9, 0xbb, 0x5d, 0xab, 0xde, 0x18, 0xe1, 0x2b, 0x2d,
	0xa6, 0x6d, 0x0d, 0xeb, 0x5d, 0x11, 0xaa, 0x12, 0x54, 0x86, 0x03, 0xed, 0xb9, 0xfb, 0x09, 0x00,
	0x00, 0xff, 0xff, 0x4e, 0x01, 0x73, 0x97, 0xa9, 0x01, 0x00, 0x00,
}

func (m *IndexedPubKey) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *IndexedPubKey) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *IndexedPubKey) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.PubKey) > 0 {
		i -= len(m.PubKey)
		copy(dAtA[i:], m.PubKey)
		i = encodeVarintPubkey(dAtA, i, uint64(len(m.PubKey)))
		i--
		dAtA[i] = 0x12
	}
	if m.Index != 0 {
		i = encodeVarintPubkey(dAtA, i, uint64(m.Index))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *ProvingScheme) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ProvingScheme) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ProvingScheme) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.ActivationHeight != 0 {
		i = encodeVarintPubkey(dAtA, i, uint64(m.ActivationHeight))
		i--
		dAtA[i] = 0x18
	}
	if m.IsActivated {
		i--
		if m.IsActivated {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x10
	}
	if m.Index != 0 {
		i = encodeVarintPubkey(dAtA, i, uint64(m.Index))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.ActivationLag != 0 {
		i = encodeVarintPubkey(dAtA, i, uint64(m.ActivationLag))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintPubkey(dAtA []byte, offset int, v uint64) int {
	offset -= sovPubkey(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *IndexedPubKey) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Index != 0 {
		n += 1 + sovPubkey(uint64(m.Index))
	}
	l = len(m.PubKey)
	if l > 0 {
		n += 1 + l + sovPubkey(uint64(l))
	}
	return n
}

func (m *ProvingScheme) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Index != 0 {
		n += 1 + sovPubkey(uint64(m.Index))
	}
	if m.IsActivated {
		n += 2
	}
	if m.ActivationHeight != 0 {
		n += 1 + sovPubkey(uint64(m.ActivationHeight))
	}
	return n
}

func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ActivationLag != 0 {
		n += 1 + sovPubkey(uint64(m.ActivationLag))
	}
	return n
}

func sovPubkey(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozPubkey(x uint64) (n int) {
	return sovPubkey(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *IndexedPubKey) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPubkey
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
			return fmt.Errorf("proto: IndexedPubKey: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: IndexedPubKey: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
			}
			m.Index = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPubkey
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
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPubkey
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
				return ErrInvalidLengthPubkey
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthPubkey
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PubKey = append(m.PubKey[:0], dAtA[iNdEx:postIndex]...)
			if m.PubKey == nil {
				m.PubKey = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPubkey(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPubkey
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
func (m *ProvingScheme) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPubkey
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
			return fmt.Errorf("proto: ProvingScheme: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ProvingScheme: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
			}
			m.Index = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPubkey
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
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IsActivated", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPubkey
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
			m.IsActivated = bool(v != 0)
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ActivationHeight", wireType)
			}
			m.ActivationHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPubkey
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ActivationHeight |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipPubkey(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPubkey
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
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPubkey
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
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ActivationLag", wireType)
			}
			m.ActivationLag = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPubkey
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ActivationLag |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipPubkey(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPubkey
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
func skipPubkey(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowPubkey
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
					return 0, ErrIntOverflowPubkey
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
					return 0, ErrIntOverflowPubkey
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
				return 0, ErrInvalidLengthPubkey
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupPubkey
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthPubkey
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthPubkey        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowPubkey          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupPubkey = fmt.Errorf("proto: unexpected end of group")
)
