// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sedachain/wasm_storage/v1/genesis.proto

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

// GenesisState defines wasm-storage module's genesis state.
type GenesisState struct {
	Params               Params          `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	OraclePrograms       []OracleProgram `protobuf:"bytes,2,rep,name=oracle_programs,json=oraclePrograms,proto3" json:"oracle_programs"`
	ExecutorWasms        []ExecutorWasm  `protobuf:"bytes,3,rep,name=executor_wasms,json=executorWasms,proto3" json:"executor_wasms"`
	CoreContractRegistry string          `protobuf:"bytes,4,opt,name=core_contract_registry,json=coreContractRegistry,proto3" json:"core_contract_registry,omitempty"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_f7dee811c18d199f, []int{0}
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

func (m *GenesisState) GetOraclePrograms() []OracleProgram {
	if m != nil {
		return m.OraclePrograms
	}
	return nil
}

func (m *GenesisState) GetExecutorWasms() []ExecutorWasm {
	if m != nil {
		return m.ExecutorWasms
	}
	return nil
}

func (m *GenesisState) GetCoreContractRegistry() string {
	if m != nil {
		return m.CoreContractRegistry
	}
	return ""
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "sedachain.wasm_storage.v1.GenesisState")
}

func init() {
	proto.RegisterFile("sedachain/wasm_storage/v1/genesis.proto", fileDescriptor_f7dee811c18d199f)
}

var fileDescriptor_f7dee811c18d199f = []byte{
	// 328 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x91, 0x41, 0x4b, 0x02, 0x41,
	0x14, 0xc7, 0x77, 0x55, 0x84, 0xd6, 0x32, 0x58, 0x24, 0xcc, 0xc3, 0x66, 0x5d, 0xdc, 0x43, 0xee,
	0xa2, 0x05, 0x1d, 0x03, 0x23, 0x3a, 0x66, 0x16, 0x08, 0x5d, 0x96, 0x71, 0x7a, 0x8c, 0x0b, 0xea,
	0x5b, 0xe6, 0x8d, 0xa6, 0xdf, 0xa2, 0x0f, 0xd3, 0x87, 0xf0, 0xe8, 0xb1, 0x53, 0x84, 0x7e, 0x91,
	0xd8, 0xd9, 0x35, 0xb6, 0x83, 0xde, 0x66, 0xde, 0xfb, 0xff, 0x7e, 0xf3, 0x98, 0x67, 0x35, 0x08,
	0xde, 0x18, 0x1f, 0xb2, 0x70, 0xe2, 0xbf, 0x33, 0x1a, 0x07, 0xa4, 0x50, 0x32, 0x01, 0xfe, 0xac,
	0xe5, 0x0b, 0x98, 0x00, 0x85, 0xe4, 0x45, 0x12, 0x15, 0xda, 0xa7, 0x7f, 0x41, 0x2f, 0x1b, 0xf4,
	0x66, 0xad, 0x5a, 0x45, 0xa0, 0x40, 0x9d, 0xf2, 0xe3, 0x53, 0x02, 0xd4, 0x2e, 0x77, 0x9b, 0xff,
	0x09, 0x74, 0xfa, 0xe2, 0x33, 0x67, 0x1d, 0x3e, 0x24, 0x0f, 0x3e, 0x2b, 0xa6, 0xc0, 0xbe, 0xb5,
	0x8a, 0x11, 0x93, 0x6c, 0x4c, 0x55, 0xb3, 0x6e, 0xba, 0xa5, 0xf6, 0xb9, 0xb7, 0x73, 0x00, 0xaf,
	0xab, 0x83, 0x9d, 0xc2, 0xf2, 0xfb, 0xcc, 0xe8, 0xa5, 0x98, 0xdd, 0xb7, 0x8e, 0x51, 0x32, 0x3e,
	0x82, 0x20, 0x92, 0x28, 0xb4, 0x29, 0x57, 0xcf, 0xbb, 0xa5, 0xb6, 0xbb, 0xc7, 0xf4, 0xa8, 0x89,
	0x6e, 0x02, 0xa4, 0xc2, 0x32, 0x66, 0x8b, 0x64, 0xbf, 0x58, 0x65, 0x98, 0x03, 0x9f, 0x2a, 0x94,
	0x41, 0xcc, 0x53, 0x35, 0xaf, 0xbd, 0x8d, 0x3d, 0xde, 0xfb, 0x14, 0xe8, 0x33, 0xda, 0x6a, 0x8f,
	0x20, 0x53, 0x23, 0xfb, 0xda, 0x3a, 0xe1, 0x28, 0x21, 0xe0, 0x38, 0x51, 0x92, 0x71, 0x15, 0x48,
	0x10, 0x21, 0x29, 0xb9, 0xa8, 0x16, 0xea, 0xa6, 0x7b, 0xd0, 0xab, 0xc4, 0xdd, 0xbb, 0xb4, 0xd9,
	0x4b, 0x7b, 0x9d, 0xa7, 0xe5, 0xda, 0x31, 0x57, 0x6b, 0xc7, 0xfc, 0x59, 0x3b, 0xe6, 0xc7, 0xc6,
	0x31, 0x56, 0x1b, 0xc7, 0xf8, 0xda, 0x38, 0xc6, 0xeb, 0x8d, 0x08, 0xd5, 0x70, 0x3a, 0xf0, 0x38,
	0x8e, 0xfd, 0x78, 0x2e, 0xfd, 0xcd, 0x1c, 0x47, 0xfa, 0xd2, 0x4c, 0xf6, 0x32, 0xd7, 0x9b, 0x68,
	0x6e, 0x37, 0xa3, 0x16, 0x11, 0xd0, 0xa0, 0xa8, 0x93, 0x57, 0xbf, 0x01, 0x00, 0x00, 0xff, 0xff,
	0x44, 0x64, 0x2c, 0x95, 0x1a, 0x02, 0x00, 0x00,
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
	if len(m.CoreContractRegistry) > 0 {
		i -= len(m.CoreContractRegistry)
		copy(dAtA[i:], m.CoreContractRegistry)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.CoreContractRegistry)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.ExecutorWasms) > 0 {
		for iNdEx := len(m.ExecutorWasms) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ExecutorWasms[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.OraclePrograms) > 0 {
		for iNdEx := len(m.OraclePrograms) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.OraclePrograms[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.OraclePrograms) > 0 {
		for _, e := range m.OraclePrograms {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.ExecutorWasms) > 0 {
		for _, e := range m.ExecutorWasms {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = len(m.CoreContractRegistry)
	if l > 0 {
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
				return fmt.Errorf("proto: wrong wireType = %d for field OraclePrograms", wireType)
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
			m.OraclePrograms = append(m.OraclePrograms, OracleProgram{})
			if err := m.OraclePrograms[len(m.OraclePrograms)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExecutorWasms", wireType)
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
			m.ExecutorWasms = append(m.ExecutorWasms, ExecutorWasm{})
			if err := m.ExecutorWasms[len(m.ExecutorWasms)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CoreContractRegistry", wireType)
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
			m.CoreContractRegistry = string(dAtA[iNdEx:postIndex])
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
