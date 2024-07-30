package keeper_test

import (
	"os"
	"path/filepath"

	"cosmossdk.io/collections"
	gomock "go.uber.org/mock/gomock"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

func (s *KeeperTestSuite) TestMsgServer_AddKey() {
	tests := []struct {
		name    string
		msg     *types.MsgAddKey
		want    *types.MsgAddKeyResponse
		wantErr error
	}{
		{
			name: "Happy Path",
			msg: &types.MsgAddKey{
				ValidatorAddress: "sedavaloper10hpwdkc76wgqm5lg4my6vz33kps0jr05u9uxga",
				Index:            0,
				Pubkey: func() *codectypes.Any {
					pk, err := utils.InitializeVRFKey(s.serverCtx.Config, "", "")
					s.Require().NoError(err)
					pkAny, err := codectypes.NewAnyWithValue(pk)
					s.Require().NoError(err)
					return pkAny
				}(),
			},
			want:    &types.MsgAddKeyResponse{},
			wantErr: nil,
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
			s.mockStakingKeeper.EXPECT().GetValidator(gomock.Any(), gomock.Any()).Return(stakingtypes.Validator{}, nil)

			got, err := s.msgSrvr.AddKey(s.ctx, tt.msg)
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Nil(got)
				return
			}
			s.Require().NoError(err)
			s.Require().NotNil(got)

			valAddr, err := s.valCdc.StringToBytes(tt.msg.ValidatorAddress)
			s.Require().NoError(err)

			pkActual, err := s.keeper.PubKeys.Get(s.ctx, collections.Join(valAddr, tt.msg.Index))
			pkExpected, err := tt.msg.PublicKey()
			s.Require().NoError(err)

			s.Require().NoError(err)
			s.Require().Equal(pkExpected, pkActual)
		})
	}
}
