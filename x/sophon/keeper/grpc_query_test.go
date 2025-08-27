package keeper_test

import (
	"encoding/hex"

	"cosmossdk.io/math"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

func (s *KeeperTestSuite) TestQuerier_SophonInfo() {
	pubKeyHex := "041b84c5567b126440995d3ed5aaba0565d71e1834604819ff9c17f5e9d5dd078f70beaf8f588b541507fed6a642c5ab42dfdf8120a7f639de5122d47a69a8e8d1"
	pubKey, err := hex.DecodeString(pubKeyHex)
	s.Require().NoError(err)

	tests := []struct {
		name       string
		sophonInfo *types.SophonInfo
		pubKeyHex  string
		wantErr    error
	}{
		{
			name: "Existing sophon",
			sophonInfo: &types.SophonInfo{
				Id:           0,
				OwnerAddress: "owner",
				AdminAddress: "admin",
				Address:      "address",
				PublicKey:    pubKey,
				Memo:         "memo",
				Balance:      math.NewInt(0),
				UsedCredits:  math.NewInt(0),
			},
			pubKeyHex: "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			wantErr:   nil,
		},
		{
			name:      "Unknown sophon",
			pubKeyHex: "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			wantErr:   sdkerrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.sophonInfo != nil {
				pubkeyBytes, err := hex.DecodeString(tt.pubKeyHex)
				s.Require().NoError(err)

				err = s.keeper.SetSophonInfo(s.ctx, pubkeyBytes, *tt.sophonInfo)
				s.Require().NoError(err)
			}

			res, err := s.queryClient.SophonInfo(s.ctx, &types.QuerySophonInfoRequest{SophonPubKey: tt.pubKeyHex})
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Nil(res)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(res)
			s.Require().Equal(tt.sophonInfo, res.Info)
		})
	}
}
