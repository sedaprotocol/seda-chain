package testutil

import (
	"reflect"

	gomock "go.uber.org/mock/gomock"
	protov2 "google.golang.org/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ValidTx   = []byte("valid")
	InvalidTx = []byte("invalid")
	LargeTx   = []byte("large")
)

type MockTx struct {
	gas uint64
}

func (m *MockTx) GetMsgs() []sdk.Msg {
	return []sdk.Msg{}
}

func (m *MockTx) GetMsgsV2() ([]protov2.Message, error) {
	return []protov2.Message{}, nil
}

func (m *MockTx) GetGas() uint64 {
	return m.gas
}

func NewMockTx(gas uint64) *MockTx {
	return &MockTx{gas: gas}
}

// MockSEDASigner is a mock of SEDASigner interface.
type MockTxVerifier struct {
	ctrl     *gomock.Controller
	recorder *MockTxVerifierMockRecorder
}

// MockTxVerifierMockRecorder is the mock recorder for MockSEDASigner.
type MockTxVerifierMockRecorder struct {
	mock *MockTxVerifier
}

// NewMockSEDASigner creates a new mock instance.
func NewMockTxVerifier(ctrl *gomock.Controller) *MockTxVerifier {
	mock := &MockTxVerifier{ctrl: ctrl}
	mock.recorder = &MockTxVerifierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTxVerifier) EXPECT() *MockTxVerifierMockRecorder {
	return m.recorder
}

func (m *MockTxVerifier) PrepareProposalVerifyTx(tx sdk.Tx) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PrepareProposalVerifyTx", tx)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockTxVerifierMockRecorder) PrepareProposalVerifyTx(input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PrepareProposalVerifyTx", reflect.TypeOf((*MockTxVerifier)(nil).PrepareProposalVerifyTx), input)
}

func (m *MockTxVerifier) ProcessProposalVerifyTx(txBz []byte) (sdk.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessProposalVerifyTx", txBz)
	ret0, _ := ret[0].(sdk.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockTxVerifierMockRecorder) ProcessProposalVerifyTx(input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessProposalVerifyTx", reflect.TypeOf((*MockTxVerifier)(nil).ProcessProposalVerifyTx), input)
}

func (m *MockTxVerifier) TxDecode(txBz []byte) (sdk.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TxDecode", txBz)
	ret0, _ := ret[0].(sdk.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockTxVerifierMockRecorder) TxDecode(input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TxDecode", reflect.TypeOf((*MockTxVerifier)(nil).TxDecode), input)
}

func (m *MockTxVerifier) TxEncode(tx sdk.Tx) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TxEncode", tx)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockTxVerifierMockRecorder) TxEncode(input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TxEncode", reflect.TypeOf((*MockTxVerifier)(nil).TxEncode), input)
}
