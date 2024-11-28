// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sedachain/batching/v1/genesis.proto

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

// GenesisState defines the batching module's genesis state.
type GenesisState struct {
	// current_batch_number is the batch number of the most recently-
	// created batch.
	CurrentBatchNumber uint64            `protobuf:"varint,1,opt,name=current_batch_number,json=currentBatchNumber,proto3" json:"current_batch_number,omitempty"`
	Batches            []Batch           `protobuf:"bytes,2,rep,name=batches,proto3" json:"batches"`
	BatchData          []BatchData       `protobuf:"bytes,3,rep,name=batch_data,json=batchData,proto3" json:"batch_data"`
	DataResults        []DataResult      `protobuf:"bytes,4,rep,name=data_results,json=dataResults,proto3" json:"data_results"`
	BatchAssignments   []BatchAssignment `protobuf:"bytes,5,rep,name=batch_assignments,json=batchAssignments,proto3" json:"batch_assignments"`
	Params             Params            `protobuf:"bytes,6,opt,name=params,proto3" json:"params"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_eccca5d98d3cb479, []int{0}
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

func (m *GenesisState) GetCurrentBatchNumber() uint64 {
	if m != nil {
		return m.CurrentBatchNumber
	}
	return 0
}

func (m *GenesisState) GetBatches() []Batch {
	if m != nil {
		return m.Batches
	}
	return nil
}

func (m *GenesisState) GetBatchData() []BatchData {
	if m != nil {
		return m.BatchData
	}
	return nil
}

func (m *GenesisState) GetDataResults() []DataResult {
	if m != nil {
		return m.DataResults
	}
	return nil
}

func (m *GenesisState) GetBatchAssignments() []BatchAssignment {
	if m != nil {
		return m.BatchAssignments
	}
	return nil
}

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

// BatchAssignment represents a batch assignment for genesis export
// and import.
type BatchAssignment struct {
	BatchNumber   uint64 `protobuf:"varint,1,opt,name=batch_number,json=batchNumber,proto3" json:"batch_number,omitempty"`
	DataRequestId string `protobuf:"bytes,2,opt,name=data_request_id,json=dataRequestId,proto3" json:"data_request_id,omitempty"`
}

func (m *BatchAssignment) Reset()         { *m = BatchAssignment{} }
func (m *BatchAssignment) String() string { return proto.CompactTextString(m) }
func (*BatchAssignment) ProtoMessage()    {}
func (*BatchAssignment) Descriptor() ([]byte, []int) {
	return fileDescriptor_eccca5d98d3cb479, []int{1}
}
func (m *BatchAssignment) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BatchAssignment) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BatchAssignment.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BatchAssignment) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BatchAssignment.Merge(m, src)
}
func (m *BatchAssignment) XXX_Size() int {
	return m.Size()
}
func (m *BatchAssignment) XXX_DiscardUnknown() {
	xxx_messageInfo_BatchAssignment.DiscardUnknown(m)
}

var xxx_messageInfo_BatchAssignment proto.InternalMessageInfo

func (m *BatchAssignment) GetBatchNumber() uint64 {
	if m != nil {
		return m.BatchNumber
	}
	return 0
}

func (m *BatchAssignment) GetDataRequestId() string {
	if m != nil {
		return m.DataRequestId
	}
	return ""
}

// BatchData represents a given batch's full data.
type BatchData struct {
	BatchNumber       uint64                `protobuf:"varint,1,opt,name=batch_number,json=batchNumber,proto3" json:"batch_number,omitempty"`
	DataResultEntries DataResultTreeEntries `protobuf:"bytes,2,opt,name=data_result_entries,json=dataResultEntries,proto3" json:"data_result_entries"`
	ValidatorEntries  []ValidatorTreeEntry  `protobuf:"bytes,3,rep,name=validator_entries,json=validatorEntries,proto3" json:"validator_entries"`
	BatchSignatures   []BatchSignatures     `protobuf:"bytes,4,rep,name=batch_signatures,json=batchSignatures,proto3" json:"batch_signatures"`
}

func (m *BatchData) Reset()         { *m = BatchData{} }
func (m *BatchData) String() string { return proto.CompactTextString(m) }
func (*BatchData) ProtoMessage()    {}
func (*BatchData) Descriptor() ([]byte, []int) {
	return fileDescriptor_eccca5d98d3cb479, []int{2}
}
func (m *BatchData) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BatchData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BatchData.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BatchData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BatchData.Merge(m, src)
}
func (m *BatchData) XXX_Size() int {
	return m.Size()
}
func (m *BatchData) XXX_DiscardUnknown() {
	xxx_messageInfo_BatchData.DiscardUnknown(m)
}

var xxx_messageInfo_BatchData proto.InternalMessageInfo

func (m *BatchData) GetBatchNumber() uint64 {
	if m != nil {
		return m.BatchNumber
	}
	return 0
}

func (m *BatchData) GetDataResultEntries() DataResultTreeEntries {
	if m != nil {
		return m.DataResultEntries
	}
	return DataResultTreeEntries{}
}

func (m *BatchData) GetValidatorEntries() []ValidatorTreeEntry {
	if m != nil {
		return m.ValidatorEntries
	}
	return nil
}

func (m *BatchData) GetBatchSignatures() []BatchSignatures {
	if m != nil {
		return m.BatchSignatures
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "sedachain.batching.v1.GenesisState")
	proto.RegisterType((*BatchAssignment)(nil), "sedachain.batching.v1.BatchAssignment")
	proto.RegisterType((*BatchData)(nil), "sedachain.batching.v1.BatchData")
}

func init() {
	proto.RegisterFile("sedachain/batching/v1/genesis.proto", fileDescriptor_eccca5d98d3cb479)
}

var fileDescriptor_eccca5d98d3cb479 = []byte{
	// 495 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x93, 0x51, 0x8b, 0xd3, 0x40,
	0x10, 0xc7, 0x9b, 0xb6, 0x56, 0x3a, 0xad, 0xd4, 0xae, 0x27, 0x84, 0x43, 0x63, 0xae, 0xca, 0x51,
	0x41, 0x13, 0xef, 0xee, 0x51, 0x5f, 0x3c, 0x3c, 0x44, 0x41, 0x91, 0x9e, 0x28, 0xca, 0x41, 0xd8,
	0x4d, 0x96, 0x34, 0xd0, 0x26, 0x75, 0x77, 0x53, 0xbc, 0x6f, 0xe1, 0xd7, 0xf0, 0x9b, 0x9c, 0x6f,
	0xf7, 0xe8, 0x93, 0x48, 0xfb, 0x45, 0x24, 0xb3, 0x9b, 0xdc, 0x79, 0xb4, 0xd5, 0xb7, 0x64, 0xe6,
	0x3f, 0xbf, 0xff, 0xee, 0xcc, 0x0e, 0xdc, 0x97, 0x3c, 0xa2, 0xe1, 0x98, 0x26, 0xa9, 0xcf, 0xa8,
	0x0a, 0xc7, 0x49, 0x1a, 0xfb, 0xf3, 0x3d, 0x3f, 0xe6, 0x29, 0x97, 0x89, 0xf4, 0x66, 0x22, 0x53,
	0x19, 0xb9, 0x5d, 0x89, 0xbc, 0x52, 0xe4, 0xcd, 0xf7, 0xb6, 0xb7, 0xe2, 0x2c, 0xce, 0x50, 0xe1,
	0x17, 0x5f, 0x5a, 0xbc, 0xfd, 0x60, 0x35, 0xb1, 0x2a, 0x44, 0xd5, 0xe0, 0x7b, 0x03, 0xba, 0x2f,
	0xb5, 0xc9, 0xb1, 0xa2, 0x8a, 0x93, 0x27, 0xb0, 0x15, 0xe6, 0x42, 0xf0, 0x54, 0x05, 0x28, 0x0d,
	0xd2, 0x7c, 0xca, 0xb8, 0xb0, 0x2d, 0xd7, 0x1a, 0x36, 0x47, 0xc4, 0xe4, 0x0e, 0x8b, 0xd4, 0x5b,
	0xcc, 0x90, 0x67, 0x70, 0x1d, 0x95, 0x5c, 0xda, 0x75, 0xb7, 0x31, 0xec, 0xec, 0xdf, 0xf1, 0x56,
	0x9e, 0xd3, 0xc3, 0xa2, 0xc3, 0xe6, 0xd9, 0xaf, 0x7b, 0xb5, 0x51, 0x59, 0x42, 0x8e, 0x00, 0xb4,
	0x4f, 0x44, 0x15, 0xb5, 0x1b, 0x08, 0x70, 0x37, 0x01, 0x5e, 0x50, 0x45, 0x0d, 0xa4, 0xcd, 0xca,
	0x00, 0x79, 0x0d, 0xdd, 0x02, 0x10, 0x08, 0x2e, 0xf3, 0x89, 0x92, 0x76, 0x13, 0x41, 0x3b, 0x6b,
	0x40, 0x45, 0xc9, 0x08, 0x95, 0x86, 0xd4, 0x89, 0xaa, 0x88, 0x24, 0x9f, 0xa0, 0xaf, 0x8f, 0x44,
	0xa5, 0x4c, 0xe2, 0x74, 0xca, 0x53, 0x25, 0xed, 0x6b, 0x08, 0xdc, 0xdd, 0x74, 0xb2, 0xe7, 0x95,
	0xdc, 0x50, 0x6f, 0xb2, 0xbf, 0xc3, 0x92, 0x3c, 0x85, 0xd6, 0x8c, 0x0a, 0x3a, 0x95, 0x76, 0xcb,
	0xb5, 0x86, 0x9d, 0xfd, 0xbb, 0x6b, 0x78, 0xef, 0x50, 0x64, 0x30, 0xa6, 0x64, 0x70, 0x02, 0xbd,
	0x2b, 0x3e, 0x64, 0x07, 0xba, 0x2b, 0xa6, 0xd4, 0x61, 0x97, 0xc6, 0xb3, 0x0b, 0x3d, 0xd3, 0x99,
	0x2f, 0x39, 0x97, 0x2a, 0x48, 0x22, 0xbb, 0xee, 0x5a, 0xc3, 0xf6, 0xe8, 0x86, 0xbe, 0x33, 0x46,
	0x5f, 0x45, 0x83, 0x1f, 0x75, 0x68, 0x57, 0x0d, 0xfe, 0x1f, 0x30, 0x83, 0x5b, 0x97, 0x5a, 0x1e,
	0xf0, 0x54, 0x89, 0x04, 0xdf, 0x40, 0x71, 0xb1, 0x47, 0xff, 0xec, 0xfc, 0x7b, 0xc1, 0xf9, 0x91,
	0xae, 0x31, 0xf7, 0xec, 0x5f, 0x0c, 0xc1, 0x24, 0xc8, 0x09, 0xf4, 0xe7, 0x74, 0x92, 0x44, 0x54,
	0x65, 0xa2, 0x72, 0xd0, 0x8f, 0xe4, 0xe1, 0x1a, 0x87, 0x0f, 0xa5, 0xbe, 0x34, 0x38, 0x2d, 0xa7,
	0x51, 0x91, 0x4a, 0xfa, 0x47, 0xd0, 0x13, 0x0a, 0x8a, 0x7e, 0x52, 0x95, 0x0b, 0x5e, 0x3e, 0x9c,
	0x8d, 0x73, 0x3e, 0xae, 0xd4, 0x86, 0xdc, 0x63, 0x57, 0xc2, 0x6f, 0xce, 0x16, 0x8e, 0x75, 0xbe,
	0x70, 0xac, 0xdf, 0x0b, 0xc7, 0xfa, 0xb6, 0x74, 0x6a, 0xe7, 0x4b, 0xa7, 0xf6, 0x73, 0xe9, 0xd4,
	0x3e, 0x1f, 0xc4, 0x89, 0x1a, 0xe7, 0xcc, 0x0b, 0xb3, 0xa9, 0x5f, 0x58, 0xe0, 0x16, 0x86, 0xd9,
	0x04, 0x7f, 0x1e, 0xeb, 0x75, 0xfd, 0x7a, 0xb1, 0xb0, 0xea, 0x74, 0xc6, 0x25, 0x6b, 0xa1, 0xea,
	0xe0, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xaf, 0xeb, 0x98, 0x25, 0x25, 0x04, 0x00, 0x00,
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
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x32
	if len(m.BatchAssignments) > 0 {
		for iNdEx := len(m.BatchAssignments) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.BatchAssignments[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x2a
		}
	}
	if len(m.DataResults) > 0 {
		for iNdEx := len(m.DataResults) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DataResults[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.BatchData) > 0 {
		for iNdEx := len(m.BatchData) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.BatchData[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.Batches) > 0 {
		for iNdEx := len(m.Batches) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Batches[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if m.CurrentBatchNumber != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.CurrentBatchNumber))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *BatchAssignment) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BatchAssignment) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BatchAssignment) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.DataRequestId) > 0 {
		i -= len(m.DataRequestId)
		copy(dAtA[i:], m.DataRequestId)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.DataRequestId)))
		i--
		dAtA[i] = 0x12
	}
	if m.BatchNumber != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.BatchNumber))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *BatchData) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BatchData) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BatchData) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.BatchSignatures) > 0 {
		for iNdEx := len(m.BatchSignatures) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.BatchSignatures[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.ValidatorEntries) > 0 {
		for iNdEx := len(m.ValidatorEntries) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ValidatorEntries[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	{
		size, err := m.DataResultEntries.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if m.BatchNumber != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.BatchNumber))
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
	if m.CurrentBatchNumber != 0 {
		n += 1 + sovGenesis(uint64(m.CurrentBatchNumber))
	}
	if len(m.Batches) > 0 {
		for _, e := range m.Batches {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.BatchData) > 0 {
		for _, e := range m.BatchData {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.DataResults) > 0 {
		for _, e := range m.DataResults {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.BatchAssignments) > 0 {
		for _, e := range m.BatchAssignments {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func (m *BatchAssignment) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.BatchNumber != 0 {
		n += 1 + sovGenesis(uint64(m.BatchNumber))
	}
	l = len(m.DataRequestId)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	return n
}

func (m *BatchData) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.BatchNumber != 0 {
		n += 1 + sovGenesis(uint64(m.BatchNumber))
	}
	l = m.DataResultEntries.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.ValidatorEntries) > 0 {
		for _, e := range m.ValidatorEntries {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.BatchSignatures) > 0 {
		for _, e := range m.BatchSignatures {
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
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CurrentBatchNumber", wireType)
			}
			m.CurrentBatchNumber = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CurrentBatchNumber |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Batches", wireType)
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
			m.Batches = append(m.Batches, Batch{})
			if err := m.Batches[len(m.Batches)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BatchData", wireType)
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
			m.BatchData = append(m.BatchData, BatchData{})
			if err := m.BatchData[len(m.BatchData)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataResults", wireType)
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
			m.DataResults = append(m.DataResults, DataResult{})
			if err := m.DataResults[len(m.DataResults)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BatchAssignments", wireType)
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
			m.BatchAssignments = append(m.BatchAssignments, BatchAssignment{})
			if err := m.BatchAssignments[len(m.BatchAssignments)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
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
func (m *BatchAssignment) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: BatchAssignment: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BatchAssignment: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BatchNumber", wireType)
			}
			m.BatchNumber = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BatchNumber |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataRequestId", wireType)
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
			m.DataRequestId = string(dAtA[iNdEx:postIndex])
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
func (m *BatchData) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: BatchData: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BatchData: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BatchNumber", wireType)
			}
			m.BatchNumber = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BatchNumber |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataResultEntries", wireType)
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
			if err := m.DataResultEntries.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorEntries", wireType)
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
			m.ValidatorEntries = append(m.ValidatorEntries, ValidatorTreeEntry{})
			if err := m.ValidatorEntries[len(m.ValidatorEntries)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BatchSignatures", wireType)
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
			m.BatchSignatures = append(m.BatchSignatures, BatchSignatures{})
			if err := m.BatchSignatures[len(m.BatchSignatures)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
