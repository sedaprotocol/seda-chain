package keeper_test

import (
	"os"
	"path/filepath"

	"cosmossdk.io/collections"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

func (s *KeeperTestSuite) Test_msgServer_AddVrfKey() {
	tests := []struct {
		name    string
		msg     *types.MsgAddVRFKey
		want    *types.MsgAddVrfKeyResponse
		wantErr error
	}{
		{
			name: "Happy Path",
			msg: &types.MsgAddVRFKey{
				Name:        "valid_name",
				Application: "pkr",
				Pubkey: func() *codectypes.Any {
					pk, err := utils.InitializeVRFKey(s.serverCtx.Config, "valid_name")
					s.Require().NoError(err)
					pkAny, err := codectypes.NewAnyWithValue(pk)
					s.Require().NoError(err)
					return pkAny
				}(),
			},
			want:    &types.MsgAddVrfKeyResponse{},
			wantErr: nil,
		},
		{
			name: "Validation Failed - Invalid Name",
			msg: &types.MsgAddVRFKey{
				Name:        "XX",
				Application: "pkr",
				Pubkey: func() *codectypes.Any {
					pk, err := utils.InitializeVRFKey(s.serverCtx.Config, "valid_name")
					s.Require().NoError(err)
					pkAny, err := codectypes.NewAnyWithValue(pk)
					s.Require().NoError(err)
					return pkAny
				}(),
			},
			want:    nil,
			wantErr: types.ErrInvalidInput,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.T().Cleanup(func() {
				path := s.serverCtx.Config.PrivValidatorKeyFile()
				path = filepath.Dir(path)
				s.Require().NoError(os.RemoveAll(path))
			})
			got, err := s.msgSrvr.AddVrfKey(s.ctx, tt.msg)
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Nil(got)
				return
			}
			s.Require().NoError(err)
			s.Require().NotNil(got)

			pkActual, err := s.keeper.PublicKeys.Get(s.ctx, collections.Join(tt.msg.Application, tt.msg.Name))
			pkExpected, err := tt.msg.PublicKey()
			s.Require().NoError(err)

			s.Require().NoError(err)
			s.Require().Equal(pkExpected, pkActual)
		})
	}
}
