// Code generated by MockGen. DO NOT EDIT.
// Source: app/abci/expected_keepers.go
//
// Generated by this command:
//
//	mockgen -source=app/abci/expected_keepers.go -package testutil -destination=app/abci/testutil/expected_keepers_mock.go
//

// Package testutil is a generated GoMock package.
package testutil

import (
	context "context"
	reflect "reflect"

	crypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	types "github.com/cosmos/cosmos-sdk/types"
	types0 "github.com/cosmos/cosmos-sdk/x/staking/types"
	utils "github.com/sedaprotocol/seda-chain/app/utils"
	types1 "github.com/sedaprotocol/seda-chain/x/batching/types"
	types2 "github.com/sedaprotocol/seda-chain/x/pubkey/types"
	gomock "go.uber.org/mock/gomock"
)

// MockBatchingKeeper is a mock of BatchingKeeper interface.
type MockBatchingKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockBatchingKeeperMockRecorder
}

// MockBatchingKeeperMockRecorder is the mock recorder for MockBatchingKeeper.
type MockBatchingKeeperMockRecorder struct {
	mock *MockBatchingKeeper
}

// NewMockBatchingKeeper creates a new mock instance.
func NewMockBatchingKeeper(ctrl *gomock.Controller) *MockBatchingKeeper {
	mock := &MockBatchingKeeper{ctrl: ctrl}
	mock.recorder = &MockBatchingKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBatchingKeeper) EXPECT() *MockBatchingKeeperMockRecorder {
	return m.recorder
}

// GetBatchForHeight mocks base method.
func (m *MockBatchingKeeper) GetBatchForHeight(ctx context.Context, height int64) (types1.Batch, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBatchForHeight", ctx, height)
	ret0, _ := ret[0].(types1.Batch)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBatchForHeight indicates an expected call of GetBatchForHeight.
func (mr *MockBatchingKeeperMockRecorder) GetBatchForHeight(ctx, height any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBatchForHeight", reflect.TypeOf((*MockBatchingKeeper)(nil).GetBatchForHeight), ctx, height)
}

// GetValidatorTreeEntry mocks base method.
func (m *MockBatchingKeeper) GetValidatorTreeEntry(ctx context.Context, batchNum uint64, valAddr types.ValAddress) (types1.ValidatorTreeEntry, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValidatorTreeEntry", ctx, batchNum, valAddr)
	ret0, _ := ret[0].(types1.ValidatorTreeEntry)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetValidatorTreeEntry indicates an expected call of GetValidatorTreeEntry.
func (mr *MockBatchingKeeperMockRecorder) GetValidatorTreeEntry(ctx, batchNum, valAddr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValidatorTreeEntry", reflect.TypeOf((*MockBatchingKeeper)(nil).GetValidatorTreeEntry), ctx, batchNum, valAddr)
}

// SetBatchSigSecp256k1 mocks base method.
func (m *MockBatchingKeeper) SetBatchSigSecp256k1(ctx context.Context, batchNum uint64, valAddr types.ValAddress, signature []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetBatchSigSecp256k1", ctx, batchNum, valAddr, signature)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetBatchSigSecp256k1 indicates an expected call of SetBatchSigSecp256k1.
func (mr *MockBatchingKeeperMockRecorder) SetBatchSigSecp256k1(ctx, batchNum, valAddr, signature any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetBatchSigSecp256k1", reflect.TypeOf((*MockBatchingKeeper)(nil).SetBatchSigSecp256k1), ctx, batchNum, valAddr, signature)
}

// MockPubKeyKeeper is a mock of PubKeyKeeper interface.
type MockPubKeyKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockPubKeyKeeperMockRecorder
}

// MockPubKeyKeeperMockRecorder is the mock recorder for MockPubKeyKeeper.
type MockPubKeyKeeperMockRecorder struct {
	mock *MockPubKeyKeeper
}

// NewMockPubKeyKeeper creates a new mock instance.
func NewMockPubKeyKeeper(ctrl *gomock.Controller) *MockPubKeyKeeper {
	mock := &MockPubKeyKeeper{ctrl: ctrl}
	mock.recorder = &MockPubKeyKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPubKeyKeeper) EXPECT() *MockPubKeyKeeperMockRecorder {
	return m.recorder
}

// GetValidatorKeyAtIndex mocks base method.
func (m *MockPubKeyKeeper) GetValidatorKeyAtIndex(ctx context.Context, validatorAddr types.ValAddress, index utils.SEDAKeyIndex) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValidatorKeyAtIndex", ctx, validatorAddr, index)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetValidatorKeyAtIndex indicates an expected call of GetValidatorKeyAtIndex.
func (mr *MockPubKeyKeeperMockRecorder) GetValidatorKeyAtIndex(ctx, validatorAddr, index any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValidatorKeyAtIndex", reflect.TypeOf((*MockPubKeyKeeper)(nil).GetValidatorKeyAtIndex), ctx, validatorAddr, index)
}

// GetValidatorKeys mocks base method.
func (m *MockPubKeyKeeper) GetValidatorKeys(ctx context.Context, validatorAddr string) (types2.ValidatorPubKeys, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValidatorKeys", ctx, validatorAddr)
	ret0, _ := ret[0].(types2.ValidatorPubKeys)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetValidatorKeys indicates an expected call of GetValidatorKeys.
func (mr *MockPubKeyKeeperMockRecorder) GetValidatorKeys(ctx, validatorAddr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValidatorKeys", reflect.TypeOf((*MockPubKeyKeeper)(nil).GetValidatorKeys), ctx, validatorAddr)
}

// MockStakingKeeper is a mock of StakingKeeper interface.
type MockStakingKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockStakingKeeperMockRecorder
}

// MockStakingKeeperMockRecorder is the mock recorder for MockStakingKeeper.
type MockStakingKeeperMockRecorder struct {
	mock *MockStakingKeeper
}

// NewMockStakingKeeper creates a new mock instance.
func NewMockStakingKeeper(ctrl *gomock.Controller) *MockStakingKeeper {
	mock := &MockStakingKeeper{ctrl: ctrl}
	mock.recorder = &MockStakingKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStakingKeeper) EXPECT() *MockStakingKeeperMockRecorder {
	return m.recorder
}

// GetPubKeyByConsAddr mocks base method.
func (m *MockStakingKeeper) GetPubKeyByConsAddr(arg0 context.Context, arg1 types.ConsAddress) (crypto.PublicKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPubKeyByConsAddr", arg0, arg1)
	ret0, _ := ret[0].(crypto.PublicKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPubKeyByConsAddr indicates an expected call of GetPubKeyByConsAddr.
func (mr *MockStakingKeeperMockRecorder) GetPubKeyByConsAddr(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPubKeyByConsAddr", reflect.TypeOf((*MockStakingKeeper)(nil).GetPubKeyByConsAddr), arg0, arg1)
}

// GetValidator mocks base method.
func (m *MockStakingKeeper) GetValidator(ctx context.Context, addr types.ValAddress) (types0.Validator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValidator", ctx, addr)
	ret0, _ := ret[0].(types0.Validator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetValidator indicates an expected call of GetValidator.
func (mr *MockStakingKeeperMockRecorder) GetValidator(ctx, addr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValidator", reflect.TypeOf((*MockStakingKeeper)(nil).GetValidator), ctx, addr)
}

// GetValidatorByConsAddr mocks base method.
func (m *MockStakingKeeper) GetValidatorByConsAddr(ctx context.Context, consAddr types.ConsAddress) (types0.Validator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValidatorByConsAddr", ctx, consAddr)
	ret0, _ := ret[0].(types0.Validator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetValidatorByConsAddr indicates an expected call of GetValidatorByConsAddr.
func (mr *MockStakingKeeperMockRecorder) GetValidatorByConsAddr(ctx, consAddr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValidatorByConsAddr", reflect.TypeOf((*MockStakingKeeper)(nil).GetValidatorByConsAddr), ctx, consAddr)
}
