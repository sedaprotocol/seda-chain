package keeper_test

import (
	"encoding/hex"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

func (s *KeeperTestSuite) TestMsgServer_RegisterSophon() {
	pubKeyHex := "02095af5db08cef43871a4aa48a80bdddc5249e4234e7432c3d7eca14f31261b10"
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
	pubKeyHex := "02095af5db08cef43871a4aa48a80bdddc5249e4234e7432c3d7eca14f31261b10"

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
		s.Require().ErrorIs(err, types.ErrSophonAlreadyExists)
	})
}

func (s *KeeperTestSuite) TestMsgServer_EditSophon() {
	pubKeyHex := "02095af5db08cef43871a4aa48a80bdddc5249e4234e7432c3d7eca14f31261b10"
	pubKey, err := hex.DecodeString(pubKeyHex)
	s.Require().NoError(err)

	tests := []struct {
		name     string
		msg      *types.MsgEditSophon
		expected *types.SophonInfo
		wantErr  error
	}{
		{
			name: "Edit admin address",
			msg: &types.MsgEditSophon{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewAdminAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				SophonPublicKey: pubKeyHex,
				NewAddress:      types.DoNotModifyField,
				NewMemo:         types.DoNotModifyField,
				NewPublicKey:    types.DoNotModifyField,
			},
			expected: &types.SophonInfo{
				Id:           0,
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    pubKey,
				Memo:         "",
				Balance:      math.NewInt(0),
				UsedCredits:  math.NewInt(0),
			},
		},
		{
			name: "Not the expected owner",
			msg: &types.MsgEditSophon{
				OwnerAddress:    "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				NewAdminAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				SophonPublicKey: pubKeyHex,
				NewAddress:      types.DoNotModifyField,
				NewMemo:         types.DoNotModifyField,
				NewPublicKey:    types.DoNotModifyField,
			},
			wantErr: sdkerrors.ErrorInvalidSigner,
		},
		{
			name: "Unknown sophon",
			msg: &types.MsgEditSophon{
				OwnerAddress:    "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				NewAdminAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				SophonPublicKey: "00",
				NewAddress:      types.DoNotModifyField,
				NewMemo:         types.DoNotModifyField,
				NewPublicKey:    types.DoNotModifyField,
			},
			wantErr: sdkerrors.ErrNotFound,
		},
		{
			name: "Edit address",
			msg: &types.MsgEditSophon{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewAdminAddress: types.DoNotModifyField,
				SophonPublicKey: pubKeyHex,
				NewAddress:      "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				NewMemo:         types.DoNotModifyField,
				NewPublicKey:    types.DoNotModifyField,
			},
			expected: &types.SophonInfo{
				Id:           0,
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				PublicKey:    pubKey,
				Memo:         "",
				Balance:      math.NewInt(0),
				UsedCredits:  math.NewInt(0),
			},
		},
		{
			name: "Edit memo",
			msg: &types.MsgEditSophon{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewAdminAddress: types.DoNotModifyField,
				SophonPublicKey: pubKeyHex,
				NewAddress:      types.DoNotModifyField,
				NewMemo:         "This is a sweet sophon",
				NewPublicKey:    types.DoNotModifyField,
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
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			_, err = s.msgSrvr.RegisterSophon(s.ctx, &types.MsgRegisterSophon{
				Authority:    s.keeper.GetAuthority(),
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    pubKeyHex,
				Memo:         "",
			})
			s.Require().NoError(err)

			response, err := s.msgSrvr.EditSophon(s.ctx, test.msg)
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

	s.Run("Edit public key", func() {
		newPubKeyHex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"
		newPubKey, err := hex.DecodeString(newPubKeyHex)
		s.Require().NoError(err)

		_, err = s.msgSrvr.RegisterSophon(s.ctx, &types.MsgRegisterSophon{
			Authority:    s.keeper.GetAuthority(),
			OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			PublicKey:    pubKeyHex,
			Memo:         "",
		})
		s.Require().NoError(err)

		_, err = s.msgSrvr.EditSophon(s.ctx, &types.MsgEditSophon{
			OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewAdminAddress: types.DoNotModifyField,
			SophonPublicKey: pubKeyHex,
			NewAddress:      types.DoNotModifyField,
			NewMemo:         types.DoNotModifyField,
			NewPublicKey:    newPubKeyHex,
		})
		s.Require().NoError(err)

		// The old public key should be deleted
		_, err = s.keeper.GetSophonInfo(s.ctx, pubKey)
		s.Require().ErrorIs(err, collections.ErrNotFound)

		// The new public key should be set
		sophonInfo, err := s.keeper.GetSophonInfo(s.ctx, newPubKey)
		s.Require().NoError(err)

		s.Require().Equal(newPubKey, sophonInfo.PublicKey)
	})
}
