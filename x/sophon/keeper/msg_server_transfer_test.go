package keeper_test

import (
	"encoding/hex"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

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

func (s *KeeperTestSuite) TestMsgServer_OwnershipTransferSpecialCases() {
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

		// Transfer to a new owner
		_, err = s.msgSrvr.TransferOwnership(s.ctx, &types.MsgTransferOwnership{
			OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewOwnerAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			SophonPublicKey: pubKeyHex,
		})
		s.Require().NoError(err)

		// Transfer to different owner
		_, err = s.msgSrvr.TransferOwnership(s.ctx, &types.MsgTransferOwnership{
			OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewOwnerAddress: "seda1s5jxphgva2dvx4dw7mzy6dae7rn8vjqhhl0t7v",
			SophonPublicKey: pubKeyHex,
		})
		s.Require().NoError(err)

		// Initial transfer should have been replaced by the new transfer
		_, err = s.msgSrvr.AcceptOwnership(s.ctx, &types.MsgAcceptOwnership{
			NewOwnerAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			SophonPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)

		// Accept the transfer
		_, err = s.msgSrvr.AcceptOwnership(s.ctx, &types.MsgAcceptOwnership{
			NewOwnerAddress: "seda1s5jxphgva2dvx4dw7mzy6dae7rn8vjqhhl0t7v",
			SophonPublicKey: pubKeyHex,
		})
		s.Require().NoError(err)
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
