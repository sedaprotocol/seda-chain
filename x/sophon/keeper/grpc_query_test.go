package keeper_test

import (
	"encoding/hex"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

func (s *KeeperTestSuite) TestQuerier_SophonInfo() {
	pubKeyHex := "02095af5db08cef43871a4aa48a80bdddc5249e4234e7432c3d7eca14f31261b10"
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

func (s *KeeperTestSuite) TestQuerier_SophonTransfer() {
	pubKeyHex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"
	pubKey, err := hex.DecodeString(pubKeyHex)
	s.Require().NoError(err)

	newOwnerAddress := "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5"
	newOwnerAddressBz, err := sdk.AccAddressFromBech32(newOwnerAddress)
	s.Require().NoError(err)

	s.Run("No registered sophon", func() {
		res, err := s.queryClient.SophonTransfer(s.ctx, &types.QuerySophonTransferRequest{SophonPubKey: pubKeyHex})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
		s.Require().Nil(res)
	})

	s.Run("Pending sophon transfer", func() {
		sophonInfo, err := s.keeper.CreateSophonInfo(s.ctx, pubKey, types.SophonInputs{
			OwnerAddress: "owner",
			AdminAddress: "admin",
			Address:      "address",
			PublicKey:    pubKey,
			Memo:         "memo",
		})
		s.Require().NoError(err)

		err = s.keeper.SetSophonTransfer(s.ctx, sophonInfo.Id, newOwnerAddressBz)
		s.Require().NoError(err)
		res, err := s.queryClient.SophonTransfer(s.ctx, &types.QuerySophonTransferRequest{SophonPubKey: pubKeyHex})
		s.Require().NoError(err)
		s.Require().NotNil(res)
		s.Require().Equal(newOwnerAddress, res.NewOwnerAddress)
	})

	s.Run("No pending transfer", func() {
		_, err := s.keeper.CreateSophonInfo(s.ctx, pubKey, types.SophonInputs{
			OwnerAddress: "owner",
			AdminAddress: "admin",
			Address:      "address",
			PublicKey:    pubKey,
			Memo:         "memo",
		})
		s.Require().NoError(err)

		res, err := s.queryClient.SophonTransfer(s.ctx, &types.QuerySophonTransferRequest{SophonPubKey: pubKeyHex})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
		s.Require().Nil(res)
	})
}
