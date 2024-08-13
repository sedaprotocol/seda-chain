package keeper_test

import (
	"os"
	"path/filepath"

	"cosmossdk.io/collections"
	gomock "go.uber.org/mock/gomock"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

func (s *KeeperTestSuite) TestMsgServer_AddKey() {
	pubKeys, valAddrs := s.generatePubKeysAndValAddrs(2)

	tests := []struct {
		name         string
		msg          *types.MsgAddKey
		valAddrBytes []byte
		wantErr      error
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
				},
			},
			valAddrBytes: valAddrs[0].Bytes(),
			wantErr:      nil,
		},
		{
			name: "Invalid Any",
			msg: &types.MsgAddKey{
				ValidatorAddr: valAddrs[1].String(),
				IndexedPubKeys: []types.IndexedPubKey{
					{
						Index: 0,
						PubKey: func() *codectypes.Any {
							any, err := codectypes.NewAnyWithValue(&stakingtypes.Commission{})
							s.Require().NoError(err)
							return any
						}(),
					},
				},
			},
			valAddrBytes: valAddrs[1].Bytes(),
			wantErr:      sdkerrors.ErrInvalidType,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.T().Cleanup(func() {
				path := s.serverCtx.Config.PrivValidatorKeyFile()
				path = filepath.Dir(path)
				s.Require().NoError(os.RemoveAll(path))
			})

			// Validator must exist.
			s.mockStakingKeeper.EXPECT().GetValidator(gomock.Any(), tt.valAddrBytes).Return(stakingtypes.Validator{}, nil)

			got, err := s.msgSrvr.AddKey(s.ctx, tt.msg)
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Nil(got)
				return
			}
			s.Require().NoError(err)
			s.Require().NotNil(got)

			pkActual, err := s.keeper.PubKeys.Get(s.ctx, collections.Join(tt.valAddrBytes, tt.msg.IndexedPubKeys[0].Index))
			s.Require().NoError(err)
			pkExpected, ok := tt.msg.IndexedPubKeys[0].PubKey.GetCachedValue().(cryptotypes.PubKey)
			s.Require().True(ok)
			s.Require().Equal(pkExpected, pkActual)
		})
	}
}
