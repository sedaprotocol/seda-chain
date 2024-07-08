package keeper_test

import (
	"fmt"
	"os"
	"path/filepath"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

func (s *KeeperTestSuite) TestQuerier_KeysByApplication() {
	pubKeys := make([]cryptotypes.PubKey, 0, 3)
	pubKeysAny := make([]*codectypes.Any, 0, 3)
	for i := 0; i < 3; i++ {
		pk, err := utils.InitializeVRFKey(s.serverCtx.Config, fmt.Sprintf("valid_name-%d", i))
		s.Require().NoError(err)
		pubKeys = append(pubKeys, pk)

		pkAny, err := codectypes.NewAnyWithValue(pk)
		s.Require().NoError(err)
		pubKeysAny = append(pubKeysAny, pkAny)
	}

	application := "pkr"
	tests := []struct {
		name      string
		pubKeyAny []*codectypes.Any
		want      []*codectypes.Any
		wantErr   error
	}{
		{
			name:      "One keys by application: pkr",
			pubKeyAny: pubKeysAny[:1],
			want:      pubKeysAny[:1],
			wantErr:   nil,
		},
		{
			name:      "Two keys by application: pkr",
			pubKeyAny: pubKeysAny[:2],
			want:      pubKeysAny[:2],
			wantErr:   nil,
		},
		{
			name:      "Three keys by application: pkr",
			pubKeyAny: pubKeysAny[:3],
			want:      pubKeysAny[:3],
			wantErr:   nil,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.T().Cleanup(func() {
				path := s.serverCtx.Config.PrivValidatorKeyFile()
				path = filepath.Dir(path)
				s.Require().NoError(os.RemoveAll(path))
			})

			// Store the pubKeys
			for i, pk := range tt.pubKeyAny {
				addMsg := types.MsgAddVRFKey{
					Name:        fmt.Sprintf("valid_name-%d", i),
					Application: application,
					Pubkey:      pk,
				}
				resp, err := s.msgSrvr.AddVrfKey(s.ctx, &addMsg)
				s.Require().NoError(err)
				s.Require().NotNil(resp)
			}

			resp, err := s.queryClient.KeysByApplication(s.ctx, &types.KeysByApplicationRequest{Application: application})
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(len(tt.want), len(resp.Keys))
			for i, expected := range tt.want {
				s.Require().True(resp.Keys[i].Equal(expected))
			}
		})
	}
}
