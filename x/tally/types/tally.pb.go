// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sedachain/tally/v1/tally.proto

package types

import (
	cosmossdk_io_math "cosmossdk.io/math"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
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

// Params defines the parameters for the tally module.
type Params struct {
	// MaxTallyGasLimit is the maximum gas limit for a tally request.
	MaxTallyGasLimit uint64 `protobuf:"varint,1,opt,name=max_tally_gas_limit,json=maxTallyGasLimit,proto3" json:"max_tally_gas_limit,omitempty"`
	// FilterGasCostNone is the gas cost for a filter type none.
	FilterGasCostNone uint64 `protobuf:"varint,2,opt,name=filter_gas_cost_none,json=filterGasCostNone,proto3" json:"filter_gas_cost_none,omitempty"`
	// FilterGasCostMultiplierMode is the gas cost multiplier for a filter type
	// mode.
	FilterGasCostMultiplierMode uint64 `protobuf:"varint,3,opt,name=filter_gas_cost_multiplier_mode,json=filterGasCostMultiplierMode,proto3" json:"filter_gas_cost_multiplier_mode,omitempty"`
	// FilterGasCostMultiplierStdDev is the gas cost multiplier for a filter type
	// stddev.
	FilterGasCostMultiplierStdDev uint64 `protobuf:"varint,4,opt,name=filter_gas_cost_multiplier_std_dev,json=filterGasCostMultiplierStdDev,proto3" json:"filter_gas_cost_multiplier_std_dev,omitempty"`
	// GasCostBase is the base gas cost for a data request.
	GasCostBase uint64 `protobuf:"varint,5,opt,name=gas_cost_base,json=gasCostBase,proto3" json:"gas_cost_base,omitempty"`
	// GasCostFallback is the gas cost incurred for data request execution when
	// even basic consensus has not been reached.
	ExecutionGasCostFallback uint64 `protobuf:"varint,6,opt,name=execution_gas_cost_fallback,json=executionGasCostFallback,proto3" json:"execution_gas_cost_fallback,omitempty"`
	// BurnRatio is the ratio of the gas cost to be burned in case of reduced
	// payout scenarios.
	BurnRatio cosmossdk_io_math.LegacyDec `protobuf:"bytes,7,opt,name=burn_ratio,json=burnRatio,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"burn_ratio"`
	// MaxResultSize is the maximum size of the result of a data request in bytes.
	MaxResultSize uint32 `protobuf:"varint,8,opt,name=max_result_size,json=maxResultSize,proto3" json:"max_result_size,omitempty"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_2917df8a6808d5e2, []int{0}
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

func (m *Params) GetMaxTallyGasLimit() uint64 {
	if m != nil {
		return m.MaxTallyGasLimit
	}
	return 0
}

func (m *Params) GetFilterGasCostNone() uint64 {
	if m != nil {
		return m.FilterGasCostNone
	}
	return 0
}

func (m *Params) GetFilterGasCostMultiplierMode() uint64 {
	if m != nil {
		return m.FilterGasCostMultiplierMode
	}
	return 0
}

func (m *Params) GetFilterGasCostMultiplierStdDev() uint64 {
	if m != nil {
		return m.FilterGasCostMultiplierStdDev
	}
	return 0
}

func (m *Params) GetGasCostBase() uint64 {
	if m != nil {
		return m.GasCostBase
	}
	return 0
}

func (m *Params) GetExecutionGasCostFallback() uint64 {
	if m != nil {
		return m.ExecutionGasCostFallback
	}
	return 0
}

func (m *Params) GetMaxResultSize() uint32 {
	if m != nil {
		return m.MaxResultSize
	}
	return 0
}

func init() {
	proto.RegisterType((*Params)(nil), "sedachain.tally.v1.Params")
}

func init() { proto.RegisterFile("sedachain/tally/v1/tally.proto", fileDescriptor_2917df8a6808d5e2) }

var fileDescriptor_2917df8a6808d5e2 = []byte{
	// 460 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x92, 0x41, 0x6b, 0x13, 0x41,
	0x18, 0x86, 0xb3, 0xb6, 0x46, 0x3b, 0x12, 0xb4, 0x63, 0x0f, 0x6b, 0x83, 0x9b, 0x90, 0x83, 0x04,
	0x21, 0x59, 0x8b, 0xe0, 0xcd, 0x4b, 0x0c, 0x16, 0xa5, 0x15, 0xd9, 0xea, 0xc5, 0xcb, 0xf2, 0xed,
	0xec, 0xd7, 0xcd, 0xd0, 0x99, 0x9d, 0xb0, 0x33, 0x1b, 0x36, 0xfd, 0x15, 0xfe, 0x0c, 0x8f, 0x1e,
	0xfc, 0x11, 0x3d, 0x16, 0x4f, 0xa2, 0x50, 0x24, 0x39, 0xf8, 0x37, 0x64, 0x67, 0x96, 0x88, 0x82,
	0xbd, 0x0c, 0x33, 0xef, 0xf7, 0xbc, 0xef, 0x1c, 0xbe, 0x97, 0x04, 0x1a, 0x53, 0x60, 0x33, 0xe0,
	0x79, 0x68, 0x40, 0x88, 0x65, 0xb8, 0x38, 0x70, 0x97, 0xf1, 0xbc, 0x50, 0x46, 0x51, 0xba, 0x99,
	0x8f, 0x9d, 0xbc, 0x38, 0xd8, 0x7f, 0xc0, 0x94, 0x96, 0x4a, 0xc7, 0x96, 0x08, 0xdd, 0xc3, 0xe1,
	0xfb, 0x7b, 0x99, 0xca, 0x94, 0xd3, 0xeb, 0x5b, 0xa3, 0xee, 0x82, 0xe4, 0xb9, 0x0a, 0xed, 0xe9,
	0xa4, 0xc1, 0x8f, 0x2d, 0xd2, 0x7e, 0x0b, 0x05, 0x48, 0x4d, 0x47, 0xe4, 0xbe, 0x84, 0x2a, 0xb6,
	0xf1, 0x71, 0x06, 0x3a, 0x16, 0x5c, 0x72, 0xe3, 0x7b, 0x7d, 0x6f, 0xb8, 0x1d, 0xdd, 0x93, 0x50,
	0xbd, 0xab, 0x27, 0x87, 0xa0, 0x8f, 0x6a, 0x9d, 0x86, 0x64, 0xef, 0x94, 0x0b, 0x83, 0x85, 0x65,
	0x99, 0xd2, 0x26, 0xce, 0x55, 0x8e, 0xfe, 0x0d, 0xcb, 0xef, 0xba, 0xd9, 0x21, 0xe8, 0x17, 0x4a,
	0x9b, 0x37, 0x2a, 0x47, 0x3a, 0x25, 0xbd, 0x7f, 0x0d, 0xb2, 0x14, 0x86, 0xcf, 0x05, 0xc7, 0x22,
	0x96, 0x2a, 0x45, 0x7f, 0xcb, 0x7a, 0xbb, 0x7f, 0x79, 0x8f, 0x37, 0xcc, 0xb1, 0x4a, 0x91, 0xbe,
	0x22, 0x83, 0x6b, 0x52, 0xb4, 0x49, 0xe3, 0x14, 0x17, 0xfe, 0xb6, 0x0d, 0x7a, 0xf8, 0x9f, 0xa0,
	0x13, 0x93, 0x4e, 0x71, 0x41, 0x07, 0xa4, 0xb3, 0xc9, 0x48, 0x40, 0xa3, 0x7f, 0xd3, 0xba, 0xee,
	0x64, 0x8e, 0x9f, 0x80, 0x46, 0xfa, 0x9c, 0x74, 0xb1, 0x42, 0x56, 0x1a, 0xae, 0xf2, 0x3f, 0x3f,
	0x9e, 0x82, 0x10, 0x09, 0xb0, 0x33, 0xbf, 0x6d, 0x1d, 0xfe, 0x06, 0x69, 0xbe, 0x7a, 0xd9, 0xcc,
	0xe9, 0x7b, 0x42, 0x92, 0xb2, 0xc8, 0xe3, 0x02, 0x0c, 0x57, 0xfe, 0xad, 0xbe, 0x37, 0xdc, 0x99,
	0x3c, 0xbb, 0xb8, 0xea, 0xb5, 0xbe, 0x5f, 0xf5, 0xba, 0x6e, 0x63, 0x3a, 0x3d, 0x1b, 0x73, 0x15,
	0x4a, 0x30, 0xb3, 0xf1, 0x11, 0x66, 0xc0, 0x96, 0x53, 0x64, 0x5f, 0xbf, 0x8c, 0x48, 0xb3, 0xd0,
	0x29, 0xb2, 0x4f, 0xbf, 0x3e, 0x3f, 0xf6, 0xa2, 0x9d, 0x3a, 0x29, 0xaa, 0x83, 0xe8, 0x23, 0x72,
	0xb7, 0x5e, 0x55, 0x81, 0xba, 0x14, 0x26, 0xd6, 0xfc, 0x1c, 0xfd, 0xdb, 0x7d, 0x6f, 0xd8, 0x89,
	0x3a, 0x12, 0xaa, 0xc8, 0xaa, 0x27, 0xfc, 0x1c, 0x27, 0xaf, 0x2f, 0x56, 0x81, 0x77, 0xb9, 0x0a,
	0xbc, 0x9f, 0xab, 0xc0, 0xfb, 0xb8, 0x0e, 0x5a, 0x97, 0xeb, 0xa0, 0xf5, 0x6d, 0x1d, 0xb4, 0x3e,
	0x3c, 0xc9, 0xb8, 0x99, 0x95, 0xc9, 0x98, 0x29, 0x19, 0xd6, 0xd5, 0xb2, 0x6d, 0x60, 0x4a, 0xd8,
	0xc7, 0xc8, 0x15, 0xb1, 0x6a, 0xaa, 0x68, 0x96, 0x73, 0xd4, 0x49, 0xdb, 0x22, 0x4f, 0x7f, 0x07,
	0x00, 0x00, 0xff, 0xff, 0x80, 0xe9, 0xef, 0xfc, 0xaa, 0x02, 0x00, 0x00,
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
	if m.MaxResultSize != 0 {
		i = encodeVarintTally(dAtA, i, uint64(m.MaxResultSize))
		i--
		dAtA[i] = 0x40
	}
	{
		size := m.BurnRatio.Size()
		i -= size
		if _, err := m.BurnRatio.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintTally(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x3a
	if m.ExecutionGasCostFallback != 0 {
		i = encodeVarintTally(dAtA, i, uint64(m.ExecutionGasCostFallback))
		i--
		dAtA[i] = 0x30
	}
	if m.GasCostBase != 0 {
		i = encodeVarintTally(dAtA, i, uint64(m.GasCostBase))
		i--
		dAtA[i] = 0x28
	}
	if m.FilterGasCostMultiplierStdDev != 0 {
		i = encodeVarintTally(dAtA, i, uint64(m.FilterGasCostMultiplierStdDev))
		i--
		dAtA[i] = 0x20
	}
	if m.FilterGasCostMultiplierMode != 0 {
		i = encodeVarintTally(dAtA, i, uint64(m.FilterGasCostMultiplierMode))
		i--
		dAtA[i] = 0x18
	}
	if m.FilterGasCostNone != 0 {
		i = encodeVarintTally(dAtA, i, uint64(m.FilterGasCostNone))
		i--
		dAtA[i] = 0x10
	}
	if m.MaxTallyGasLimit != 0 {
		i = encodeVarintTally(dAtA, i, uint64(m.MaxTallyGasLimit))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintTally(dAtA []byte, offset int, v uint64) int {
	offset -= sovTally(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.MaxTallyGasLimit != 0 {
		n += 1 + sovTally(uint64(m.MaxTallyGasLimit))
	}
	if m.FilterGasCostNone != 0 {
		n += 1 + sovTally(uint64(m.FilterGasCostNone))
	}
	if m.FilterGasCostMultiplierMode != 0 {
		n += 1 + sovTally(uint64(m.FilterGasCostMultiplierMode))
	}
	if m.FilterGasCostMultiplierStdDev != 0 {
		n += 1 + sovTally(uint64(m.FilterGasCostMultiplierStdDev))
	}
	if m.GasCostBase != 0 {
		n += 1 + sovTally(uint64(m.GasCostBase))
	}
	if m.ExecutionGasCostFallback != 0 {
		n += 1 + sovTally(uint64(m.ExecutionGasCostFallback))
	}
	l = m.BurnRatio.Size()
	n += 1 + l + sovTally(uint64(l))
	if m.MaxResultSize != 0 {
		n += 1 + sovTally(uint64(m.MaxResultSize))
	}
	return n
}

func sovTally(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTally(x uint64) (n int) {
	return sovTally(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTally
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
				return fmt.Errorf("proto: wrong wireType = %d for field MaxTallyGasLimit", wireType)
			}
			m.MaxTallyGasLimit = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTally
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxTallyGasLimit |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field FilterGasCostNone", wireType)
			}
			m.FilterGasCostNone = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTally
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.FilterGasCostNone |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field FilterGasCostMultiplierMode", wireType)
			}
			m.FilterGasCostMultiplierMode = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTally
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.FilterGasCostMultiplierMode |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field FilterGasCostMultiplierStdDev", wireType)
			}
			m.FilterGasCostMultiplierStdDev = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTally
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.FilterGasCostMultiplierStdDev |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field GasCostBase", wireType)
			}
			m.GasCostBase = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTally
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.GasCostBase |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExecutionGasCostFallback", wireType)
			}
			m.ExecutionGasCostFallback = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTally
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ExecutionGasCostFallback |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BurnRatio", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTally
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
				return ErrInvalidLengthTally
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTally
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.BurnRatio.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxResultSize", wireType)
			}
			m.MaxResultSize = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTally
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxResultSize |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTally(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTally
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
func skipTally(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTally
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
					return 0, ErrIntOverflowTally
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
					return 0, ErrIntOverflowTally
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
				return 0, ErrInvalidLengthTally
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTally
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTally
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTally        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTally          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTally = fmt.Errorf("proto: unexpected end of group")
)
