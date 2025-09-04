package keeper_test

import (
	"encoding/hex"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/fast/types"
)

func (s *KeeperTestSuite) TestMsgServer_OwnershipTransferFull() {
	fastClient1InitialOwner := "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh"
	fastClient1WrongTransfer := "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5"
	fastClient1NewOwner := "seda1jd2q0mz0vzs75tp7lyuzf9064zccddgs8utjr5"
	fastClient1Hex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"
	fastClient1, err := hex.DecodeString(fastClient1Hex)
	s.Require().NoError(err)

	fastClient2InitialOwner := "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5"
	fastClient2NewOwner := "seda1xd04svzj6zj93g4eknhp6aq2yyptagcc2zeetj"
	fastClient2Hex := "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
	fastClient2, err := hex.DecodeString(fastClient2Hex)
	s.Require().NoError(err)

	// Register 2 fastClients
	_, err = s.msgSrvr.RegisterFastClient(s.ctx, &types.MsgRegisterFastClient{
		Authority:    s.keeper.GetAuthority(),
		OwnerAddress: fastClient1InitialOwner,
		AdminAddress: fastClient1InitialOwner,
		Address:      fastClient1InitialOwner,
		PublicKey:    fastClient1Hex,
		Memo:         "fastClient1",
	})
	s.Require().NoError(err)

	_, err = s.msgSrvr.RegisterFastClient(s.ctx, &types.MsgRegisterFastClient{
		Authority:    s.keeper.GetAuthority(),
		OwnerAddress: fastClient2InitialOwner,
		AdminAddress: fastClient2InitialOwner,
		Address:      fastClient2InitialOwner,
		PublicKey:    fastClient2Hex,
		Memo:         "fastClient2",
	})
	s.Require().NoError(err)

	// Transfer ownership of both fastClients
	_, err = s.msgSrvr.TransferOwnership(s.ctx, &types.MsgTransferOwnership{
		OwnerAddress:        fastClient1InitialOwner,
		NewOwnerAddress:     fastClient1WrongTransfer,
		FastClientPublicKey: fastClient1Hex,
	})
	s.Require().NoError(err)

	_, err = s.msgSrvr.TransferOwnership(s.ctx, &types.MsgTransferOwnership{
		OwnerAddress:        fastClient2InitialOwner,
		NewOwnerAddress:     fastClient2NewOwner,
		FastClientPublicKey: fastClient2Hex,
	})
	s.Require().NoError(err)

	// Cancel the transfer of fastClient1
	_, err = s.msgSrvr.CancelOwnershipTransfer(s.ctx, &types.MsgCancelOwnershipTransfer{
		OwnerAddress:        fastClient1InitialOwner,
		FastClientPublicKey: fastClient1Hex,
	})
	s.Require().NoError(err)

	// Create a new transfer for fastClient1
	_, err = s.msgSrvr.TransferOwnership(s.ctx, &types.MsgTransferOwnership{
		OwnerAddress:        fastClient1InitialOwner,
		NewOwnerAddress:     fastClient1NewOwner,
		FastClientPublicKey: fastClient1Hex,
	})
	s.Require().NoError(err)

	// Accept the transfer of both fastClients
	_, err = s.msgSrvr.AcceptOwnership(s.ctx, &types.MsgAcceptOwnership{
		NewOwnerAddress:     fastClient1NewOwner,
		FastClientPublicKey: fastClient1Hex,
	})
	s.Require().NoError(err)

	_, err = s.msgSrvr.AcceptOwnership(s.ctx, &types.MsgAcceptOwnership{
		NewOwnerAddress:     fastClient2NewOwner,
		FastClientPublicKey: fastClient2Hex,
	})
	s.Require().NoError(err)

	// Check the ownership of both fastClients
	fastClient, err := s.keeper.GetFastClient(s.ctx, fastClient1)
	s.Require().NoError(err)
	s.Require().Equal(fastClient1NewOwner, fastClient.OwnerAddress)

	fastClient, err = s.keeper.GetFastClient(s.ctx, fastClient2)
	s.Require().NoError(err)
	s.Require().Equal(fastClient2NewOwner, fastClient.OwnerAddress)

	// Verify that the old owner can no longer edit the fastClient
	_, err = s.msgSrvr.EditFastClient(s.ctx, &types.MsgEditFastClient{
		OwnerAddress:        fastClient1InitialOwner,
		NewAdminAddress:     fastClient1WrongTransfer,
		FastClientPublicKey: fastClient1Hex,
		NewAddress:          types.DoNotModifyField,
	})
	s.Require().ErrorIs(err, sdkerrors.ErrorInvalidSigner)
}

func (s *KeeperTestSuite) TestMsgServer_OwnershipTransferSpecialCases() {
	pubKeyHex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"

	s.Run("Non-existent fast client", func() {
		_, err := s.msgSrvr.TransferOwnership(s.ctx, &types.MsgTransferOwnership{
			OwnerAddress:        "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewOwnerAddress:     "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			FastClientPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
	})

	s.Run("Not the expected owner", func() {
		_, err := s.msgSrvr.RegisterFastClient(s.ctx, &types.MsgRegisterFastClient{
			Authority:    s.keeper.GetAuthority(),
			OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			PublicKey:    pubKeyHex,
			Memo:         "",
		})
		s.Require().NoError(err)

		_, err = s.msgSrvr.TransferOwnership(s.ctx, &types.MsgTransferOwnership{
			OwnerAddress:        "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			NewOwnerAddress:     "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			FastClientPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrorInvalidSigner)
	})

	s.Run("Pending transfer", func() {
		_, err := s.msgSrvr.RegisterFastClient(s.ctx, &types.MsgRegisterFastClient{
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
			OwnerAddress:        "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewOwnerAddress:     "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			FastClientPublicKey: pubKeyHex,
		})
		s.Require().NoError(err)

		// Transfer to different owner
		_, err = s.msgSrvr.TransferOwnership(s.ctx, &types.MsgTransferOwnership{
			OwnerAddress:        "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewOwnerAddress:     "seda1s5jxphgva2dvx4dw7mzy6dae7rn8vjqhhl0t7v",
			FastClientPublicKey: pubKeyHex,
		})
		s.Require().NoError(err)

		// Initial transfer should have been replaced by the new transfer
		_, err = s.msgSrvr.AcceptOwnership(s.ctx, &types.MsgAcceptOwnership{
			NewOwnerAddress:     "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			FastClientPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)

		// Accept the transfer
		_, err = s.msgSrvr.AcceptOwnership(s.ctx, &types.MsgAcceptOwnership{
			NewOwnerAddress:     "seda1s5jxphgva2dvx4dw7mzy6dae7rn8vjqhhl0t7v",
			FastClientPublicKey: pubKeyHex,
		})
		s.Require().NoError(err)
	})
}

func (s *KeeperTestSuite) TestMsgServer_OwnershipAcceptErrors() {
	pubKeyHex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"

	s.Run("Non-existent fastClient", func() {
		_, err := s.msgSrvr.AcceptOwnership(s.ctx, &types.MsgAcceptOwnership{
			NewOwnerAddress:     "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			FastClientPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
	})

	s.Run("No pending transfer", func() {
		_, err := s.msgSrvr.RegisterFastClient(s.ctx, &types.MsgRegisterFastClient{
			Authority:    s.keeper.GetAuthority(),
			OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			PublicKey:    pubKeyHex,
			Memo:         "",
		})
		s.Require().NoError(err)

		_, err = s.msgSrvr.AcceptOwnership(s.ctx, &types.MsgAcceptOwnership{
			NewOwnerAddress:     "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			FastClientPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
	})
}

func (s *KeeperTestSuite) TestMsgServer_OwnershipCancelErrors() {
	pubKeyHex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"

	s.Run("Non-existent fastClient", func() {
		_, err := s.msgSrvr.CancelOwnershipTransfer(s.ctx, &types.MsgCancelOwnershipTransfer{
			OwnerAddress:        "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			FastClientPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
	})

	s.Run("Not the expected owner", func() {
		_, err := s.msgSrvr.RegisterFastClient(s.ctx, &types.MsgRegisterFastClient{
			Authority:    s.keeper.GetAuthority(),
			OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			PublicKey:    pubKeyHex,
			Memo:         "",
		})
		s.Require().NoError(err)

		_, err = s.msgSrvr.CancelOwnershipTransfer(s.ctx, &types.MsgCancelOwnershipTransfer{
			OwnerAddress:        "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			FastClientPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrorInvalidSigner)
	})

	s.Run("No pending transfer", func() {
		_, err := s.msgSrvr.RegisterFastClient(s.ctx, &types.MsgRegisterFastClient{
			Authority:    s.keeper.GetAuthority(),
			OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			PublicKey:    pubKeyHex,
			Memo:         "",
		})
		s.Require().NoError(err)

		_, err = s.msgSrvr.CancelOwnershipTransfer(s.ctx, &types.MsgCancelOwnershipTransfer{
			OwnerAddress:        "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			FastClientPublicKey: pubKeyHex,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
	})
}
