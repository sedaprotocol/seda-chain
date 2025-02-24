// Code generated by MockGen. DO NOT EDIT.
// Source: x/staking/types/expected_keepers.go
//
// Generated by this command:
//
//	mockgen -source=x/staking/types/expected_keepers.go -package testutil -destination=x/staking/keeper/testutil/expected_keepers_mock.go
//

// Package testutil is a generated GoMock package.
package testutil

import (
	context "context"
	reflect "reflect"

	types "github.com/cosmos/cosmos-sdk/types"
	utils "github.com/sedaprotocol/seda-chain/app/utils"
	types0 "github.com/sedaprotocol/seda-chain/x/pubkey/types"
	gomock "go.uber.org/mock/gomock"
)

// MockPubKeyKeeper is a mock of PubKeyKeeper interface.
type MockPubKeyKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockPubKeyKeeperMockRecorder
	isgomock struct{}
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

// HasRegisteredKey mocks base method.
func (m *MockPubKeyKeeper) HasRegisteredKey(ctx context.Context, validatorAddr types.ValAddress, index utils.SEDAKeyIndex) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasRegisteredKey", ctx, validatorAddr, index)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasRegisteredKey indicates an expected call of HasRegisteredKey.
func (mr *MockPubKeyKeeperMockRecorder) HasRegisteredKey(ctx, validatorAddr, index any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasRegisteredKey", reflect.TypeOf((*MockPubKeyKeeper)(nil).HasRegisteredKey), ctx, validatorAddr, index)
}

// IsProvingSchemeActivated mocks base method.
func (m *MockPubKeyKeeper) IsProvingSchemeActivated(ctx context.Context, index utils.SEDAKeyIndex) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsProvingSchemeActivated", ctx, index)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsProvingSchemeActivated indicates an expected call of IsProvingSchemeActivated.
func (mr *MockPubKeyKeeperMockRecorder) IsProvingSchemeActivated(ctx, index any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsProvingSchemeActivated", reflect.TypeOf((*MockPubKeyKeeper)(nil).IsProvingSchemeActivated), ctx, index)
}

// StoreIndexedPubKeys mocks base method.
func (m *MockPubKeyKeeper) StoreIndexedPubKeys(ctx types.Context, valAddr types.ValAddress, pubKeys []types0.IndexedPubKey) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreIndexedPubKeys", ctx, valAddr, pubKeys)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreIndexedPubKeys indicates an expected call of StoreIndexedPubKeys.
func (mr *MockPubKeyKeeperMockRecorder) StoreIndexedPubKeys(ctx, valAddr, pubKeys any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreIndexedPubKeys", reflect.TypeOf((*MockPubKeyKeeper)(nil).StoreIndexedPubKeys), ctx, valAddr, pubKeys)
}
