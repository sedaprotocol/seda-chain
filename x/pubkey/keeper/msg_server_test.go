package keeper_test

import (
	"os"
	"path/filepath"

	gomock "go.uber.org/mock/gomock"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

func (s *KeeperTestSuite) TestMsgServer_AddKey() {
	pubKeys, valAddrs := s.generatePubKeysAndValAddrs(10)

	tests := []struct {
		name       string
		msg        *types.MsgAddKey
		valAddr    sdk.ValAddress
		wantErr    error
		wantErrMsg string
	}{
		{
			name: "Happy path",
			msg: &types.MsgAddKey{
				ValidatorAddr: valAddrs[0].String(),
				IndexedPubKeys: []types.IndexedPubKey{
					{
						Index:  0,
						PubKey: pubKeys[0],
					},
					{
						Index:  3,
						PubKey: pubKeys[1],
					},
					{
						Index:  57,
						PubKey: pubKeys[2],
					},
				},
			},
			valAddr: valAddrs[0],
			wantErr: nil,
		},
		{
			name: "Duplicate index",
			msg: &types.MsgAddKey{
				ValidatorAddr: valAddrs[1].String(),
				IndexedPubKeys: []types.IndexedPubKey{
					{
						Index:  12,
						PubKey: pubKeys[0],
					},
					{
						Index:  48,
						PubKey: pubKeys[1],
					},
					{
						Index:  5,
						PubKey: pubKeys[2],
					},
					{
						Index:  12,
						PubKey: pubKeys[3],
					},
					{
						Index:  50,
						PubKey: pubKeys[4],
					},
				},
			},
			valAddr:    valAddrs[1],
			wantErr:    sdkerrors.ErrInvalidRequest,
			wantErrMsg: "duplicate index at 12",
		},
		{
			name: "Invalid Any",
			msg: &types.MsgAddKey{
				ValidatorAddr: valAddrs[2].String(),
				IndexedPubKeys: []types.IndexedPubKey{
					{
						Index: 0,
						PubKey: func() *codectypes.Any {
							wrongAny, err := codectypes.NewAnyWithValue(&stakingtypes.Commission{})
							s.Require().NoError(err)
							return wrongAny
						}(),
					},
				},
			},
			valAddr: valAddrs[2],
			wantErr: sdkerrors.ErrInvalidType,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.T().Cleanup(func() {
				path := s.serverCtx.Config.PrivValidatorKeyFile()
				path = filepath.Dir(path)
				s.Require().NoError(os.RemoveAll(path))
			})

			// Mock validator store.
			s.mockStakingKeeper.EXPECT().GetValidator(gomock.Any(), tt.valAddr.Bytes()).Return(stakingtypes.Validator{}, nil).AnyTimes()

			got, err := s.msgSrvr.AddKey(s.ctx, tt.msg)
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Contains(err.Error(), tt.wantErrMsg)
				s.Require().Nil(got)
				return
			}
			s.Require().NoError(err)
			s.Require().NotNil(got)

			// Check the validator's keys at once.
			pksActual, err := s.keeper.GetValidatorKeys(s.ctx, tt.valAddr.String())
			s.Require().NoError(err)
			s.Require().Equal(tt.valAddr.String(), pksActual.ValidatorAddr)
			s.Require().Equal(len(tt.msg.IndexedPubKeys), len(pksActual.IndexedPubKeys))

			// Check each index.
			for _, indPubKey := range tt.msg.IndexedPubKeys {
				pkActual, err := s.keeper.GetValidatorKeyAtIndex(s.ctx, tt.valAddr, utils.SEDAKeyIndex(indPubKey.Index))
				s.Require().NoError(err)
				pkExpected, ok := indPubKey.PubKey.GetCachedValue().(cryptotypes.PubKey)
				s.Require().True(ok)
				s.Require().Equal(pkExpected, pkActual)
			}
		})
	}
}
