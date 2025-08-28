package keeper_test

import (
	"encoding/hex"

	"cosmossdk.io/collections"
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

func (s *KeeperTestSuite) TestMsgServer_EditSophon() {
	pubKeyHex := "041b84c5567b126440995d3ed5aaba0565d71e1834604819ff9c17f5e9d5dd078f70beaf8f588b541507fed6a642c5ab42dfdf8120a7f639de5122d47a69a8e8d1"
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

func (s *KeeperTestSuite) TestMsgServer_OwnershipTransferFull() {
	sophon1InitialOwner := "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh"
	sophon1WrongTransfer := "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5"
	sophon1NewOwner := "seda1jd2q0mz0vzs75tp7lyuzf9064zccddgs8utjr5"
	sophon1Hex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"
	sophon1, err := hex.DecodeString(sophon1Hex)
	s.Require().NoError(err)

	sophon2InitialOwner := "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5"
	sophon2NewOwner := "seda1xd04svzj6zj93g4eknhp6aq2yyptagcc2zeetj"
	sophon2Hex := "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
	sophon2, err := hex.DecodeString(sophon2Hex)
	s.Require().NoError(err)

	// Register 2 sophons
	_, err = s.msgSrvr.RegisterSophon(s.ctx, &types.MsgRegisterSophon{
		Authority:    s.keeper.GetAuthority(),
		OwnerAddress: sophon1InitialOwner,
		AdminAddress: sophon1InitialOwner,
		Address:      sophon1InitialOwner,
		PublicKey:    sophon1Hex,
		Memo:         "sophon1",
	})
	s.Require().NoError(err)

	_, err = s.msgSrvr.RegisterSophon(s.ctx, &types.MsgRegisterSophon{
		Authority:    s.keeper.GetAuthority(),
		OwnerAddress: sophon2InitialOwner,
		AdminAddress: sophon2InitialOwner,
		Address:      sophon2InitialOwner,
		PublicKey:    sophon2Hex,
		Memo:         "sophon2",
	})
	s.Require().NoError(err)

	// Transfer ownership of both sophons
	_, err = s.msgSrvr.TransferOwnership(s.ctx, &types.MsgTransferOwnership{
		OwnerAddress:    sophon1InitialOwner,
		NewOwnerAddress: sophon1WrongTransfer,
		SophonPublicKey: sophon1Hex,
	})
	s.Require().NoError(err)

	_, err = s.msgSrvr.TransferOwnership(s.ctx, &types.MsgTransferOwnership{
		OwnerAddress:    sophon2InitialOwner,
		NewOwnerAddress: sophon2NewOwner,
		SophonPublicKey: sophon2Hex,
	})
	s.Require().NoError(err)

	// Cancel the transfer of sophon1
	_, err = s.msgSrvr.CancelOwnershipTransfer(s.ctx, &types.MsgCancelOwnershipTransfer{
		OwnerAddress:    sophon1InitialOwner,
		SophonPublicKey: sophon1Hex,
	})
	s.Require().NoError(err)

	// Create a new transfer for sophon1
	_, err = s.msgSrvr.TransferOwnership(s.ctx, &types.MsgTransferOwnership{
		OwnerAddress:    sophon1InitialOwner,
		NewOwnerAddress: sophon1NewOwner,
		SophonPublicKey: sophon1Hex,
	})
	s.Require().NoError(err)

	// Accept the transfer of both sophons
	_, err = s.msgSrvr.AcceptOwnership(s.ctx, &types.MsgAcceptOwnership{
		NewOwnerAddress: sophon1NewOwner,
		SophonPublicKey: sophon1Hex,
	})
	s.Require().NoError(err)

	_, err = s.msgSrvr.AcceptOwnership(s.ctx, &types.MsgAcceptOwnership{
		NewOwnerAddress: sophon2NewOwner,
		SophonPublicKey: sophon2Hex,
	})
	s.Require().NoError(err)

	// Check the ownership of both sophons
	sophonInfo, err := s.keeper.GetSophonInfo(s.ctx, sophon1)
	s.Require().NoError(err)
	s.Require().Equal(sophon1NewOwner, sophonInfo.OwnerAddress)

	sophonInfo, err = s.keeper.GetSophonInfo(s.ctx, sophon2)
	s.Require().NoError(err)
	s.Require().Equal(sophon2NewOwner, sophonInfo.OwnerAddress)

	// Verify that the old owner can no longer edit the sophon
	_, err = s.msgSrvr.EditSophon(s.ctx, &types.MsgEditSophon{
		OwnerAddress:    sophon1InitialOwner,
		NewAdminAddress: sophon1WrongTransfer,
		SophonPublicKey: sophon1Hex,
		NewAddress:      types.DoNotModifyField,
	})
	s.Require().ErrorIs(err, sdkerrors.ErrorInvalidSigner)
}

func (s *KeeperTestSuite) TestMsgServer_OwnershipTransferErrors() {
	pubKeyHex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"

	s.Run("Non-existent sophon", func() {
		_, err := s.msgSrvr.TransferOwnership(s.ctx, &types.MsgTransferOwnership{
			OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewOwnerAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			SophonPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
	})

	s.Run("Not the expected owner", func() {
		_, err := s.msgSrvr.RegisterSophon(s.ctx, &types.MsgRegisterSophon{
			Authority:    s.keeper.GetAuthority(),
			OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			PublicKey:    pubKeyHex,
			Memo:         "",
		})
		s.Require().NoError(err)

		_, err = s.msgSrvr.TransferOwnership(s.ctx, &types.MsgTransferOwnership{
			OwnerAddress:    "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			NewOwnerAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			SophonPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrorInvalidSigner)
	})

	s.Run("Pending transfer", func() {
		_, err := s.msgSrvr.RegisterSophon(s.ctx, &types.MsgRegisterSophon{
			Authority:    s.keeper.GetAuthority(),
			OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			PublicKey:    pubKeyHex,
			Memo:         "",
		})
		s.Require().NoError(err)

		_, err = s.msgSrvr.TransferOwnership(s.ctx, &types.MsgTransferOwnership{
			OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewOwnerAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			SophonPublicKey: pubKeyHex,
		})
		s.Require().NoError(err)

		_, err = s.msgSrvr.TransferOwnership(s.ctx, &types.MsgTransferOwnership{
			OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewOwnerAddress: "seda1s5jxphgva2dvx4dw7mzy6dae7rn8vjqhhl0t7v",
			SophonPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrInvalidRequest)
	})
}

func (s *KeeperTestSuite) TestMsgServer_OwnershipAcceptErrors() {
	pubKeyHex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"

	s.Run("Non-existent sophon", func() {
		_, err := s.msgSrvr.AcceptOwnership(s.ctx, &types.MsgAcceptOwnership{
			NewOwnerAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			SophonPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
	})

	s.Run("No pending transfer", func() {
		_, err := s.msgSrvr.RegisterSophon(s.ctx, &types.MsgRegisterSophon{
			Authority:    s.keeper.GetAuthority(),
			OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			PublicKey:    pubKeyHex,
			Memo:         "",
		})
		s.Require().NoError(err)

		_, err = s.msgSrvr.AcceptOwnership(s.ctx, &types.MsgAcceptOwnership{
			NewOwnerAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			SophonPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
	})
}

func (s *KeeperTestSuite) TestMsgServer_OwnershipCancelErrors() {
	pubKeyHex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"

	s.Run("Non-existent sophon", func() {
		_, err := s.msgSrvr.CancelOwnershipTransfer(s.ctx, &types.MsgCancelOwnershipTransfer{
			OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			SophonPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
	})

	s.Run("Not the expected owner", func() {
		_, err := s.msgSrvr.RegisterSophon(s.ctx, &types.MsgRegisterSophon{
			Authority:    s.keeper.GetAuthority(),
			OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			PublicKey:    pubKeyHex,
			Memo:         "",
		})
		s.Require().NoError(err)

		_, err = s.msgSrvr.CancelOwnershipTransfer(s.ctx, &types.MsgCancelOwnershipTransfer{
			OwnerAddress:    "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			SophonPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrorInvalidSigner)
	})

	s.Run("No pending transfer", func() {
		_, err := s.msgSrvr.RegisterSophon(s.ctx, &types.MsgRegisterSophon{
			Authority:    s.keeper.GetAuthority(),
			OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			PublicKey:    pubKeyHex,
			Memo:         "",
		})
		s.Require().NoError(err)

		_, err = s.msgSrvr.CancelOwnershipTransfer(s.ctx, &types.MsgCancelOwnershipTransfer{
			OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			SophonPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
	})
}
