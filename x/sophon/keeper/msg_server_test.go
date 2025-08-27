package keeper_test

import (
	"encoding/hex"

	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

func (s *KeeperTestSuite) TestMsgServer_RegisterSophon() {
	pubKeyHex := "041b84c5567b126440995d3ed5aaba0565d71e1834604819ff9c17f5e9d5dd078f70beaf8f588b541507fed6a642c5ab42dfdf8120a7f639de5122d47a69a8e8d1"
	pubKey, err := hex.DecodeString(pubKeyHex)
	s.Require().NoError(err)

	tests := []struct {
		name      string
		msg       *types.MsgRegisterSophon
		expected  *types.SophonInfo
		wantErr   error
		mockSetup func()
	}{
		{
			// Should be the first test case so the ID is 0
			name: "Happy path",
			msg: &types.MsgRegisterSophon{
				Authority:    s.keeper.GetAuthority(),
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    pubKeyHex,
				Memo:         "This is a sweet sophon",
			},
			expected: &types.SophonInfo{
				Id:           0,
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    pubKey,
				Memo:         "This is a sweet sophon",
				Balance:      math.NewInt(0),
				UsedCredits:  math.NewInt(0),
			},
		},
		{
			name: "Not the expected authority",
			msg: &types.MsgRegisterSophon{
				Authority:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z6",
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    pubKeyHex,
				Memo:         "",
			},
			expected:  nil,
			wantErr:   sdkerrors.ErrorInvalidSigner,
			mockSetup: func() {},
		},
		{
			name: "Invalid public key",
			msg: &types.MsgRegisterSophon{
				Authority:    s.keeper.GetAuthority(),
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    "not a valid hex string",
				Memo:         "",
			},
			expected:  nil,
			wantErr:   sdkerrors.ErrInvalidRequest,
			mockSetup: func() {},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			response, err := s.msgSrvr.RegisterSophon(s.ctx, test.msg)
			if test.wantErr != nil {
				s.Require().ErrorIs(err, test.wantErr)
				s.Require().Nil(response)
				return
			}

			s.Require().NoError(err)
			sophonInfo, err := s.keeper.GetSophonInfo(s.ctx, pubKey)
			s.Require().NoError(err)

			s.Require().Equal(test.expected, &sophonInfo)
		})
	}
}

func (s *KeeperTestSuite) TestMsgServer_RegisterSophon_AlreadyExists() {
	pubKeyHex := "041b84c5567b126440995d3ed5aaba0565d71e1834604819ff9c17f5e9d5dd078f70beaf8f588b541507fed6a642c5ab42dfdf8120a7f639de5122d47a69a8e8d1"

	s.Run("Sophon already exists", func() {
		_, err := s.msgSrvr.RegisterSophon(s.ctx, &types.MsgRegisterSophon{
			Authority:    s.keeper.GetAuthority(),
			OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			PublicKey:    pubKeyHex,
			Memo:         "",
		})
		s.Require().NoError(err)

		_, err = s.msgSrvr.RegisterSophon(s.ctx, &types.MsgRegisterSophon{
			Authority:    s.keeper.GetAuthority(),
			OwnerAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			AdminAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			Address:      "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			PublicKey:    pubKeyHex,
			Memo:         "different memo same pubkey",
		})
		s.Require().ErrorIs(err, types.ErrAlreadyExists)
	})
}
