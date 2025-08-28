package types_test

import (
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

func (s *TypesTestSuite) TestMsgRegisterSophon_ValidateBasic() {
	pubKeyHex := "041b84c5567b126440995d3ed5aaba0565d71e1834604819ff9c17f5e9d5dd078f70beaf8f588b541507fed6a642c5ab42dfdf8120a7f639de5122d47a69a8e8d1"

	tests := []struct {
		name    string
		msg     *types.MsgRegisterSophon
		wantErr error
	}{
		{
			name: "valid",
			msg: &types.MsgRegisterSophon{
				Authority:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    pubKeyHex,
				Memo:         "This is a sweet sophon",
			},
		},
		{
			name: "invalid authority",
			msg: &types.MsgRegisterSophon{
				Authority:    "invalid",
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    pubKeyHex,
				Memo:         "This is a sweet sophon",
			},
			wantErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid owner address",
			msg: &types.MsgRegisterSophon{
				Authority:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				OwnerAddress: "invalid",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    pubKeyHex,
				Memo:         "This is a sweet sophon",
			},
			wantErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid admin address",
			msg: &types.MsgRegisterSophon{
				Authority:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "invalid",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    pubKeyHex,
				Memo:         "This is a sweet sophon",
			},
			wantErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid address",
			msg: &types.MsgRegisterSophon{
				Authority:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "invalid",
				PublicKey:    pubKeyHex,
				Memo:         "This is a sweet sophon",
			},
			wantErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "empty public key",
			msg: &types.MsgRegisterSophon{
				Authority:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    "",
				Memo:         "This is a sweet sophon",
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid public key",
			msg: &types.MsgRegisterSophon{
				Authority:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    "not a valid hex string",
				Memo:         "This is a sweet sophon",
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid memo",
			msg: &types.MsgRegisterSophon{
				Authority:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    pubKeyHex,
				Memo:         strings.Repeat("a", 3001),
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := test.msg.ValidateBasic()

			if test.wantErr != nil {
				s.Require().ErrorIs(err, test.wantErr)
				return
			}

			s.Require().NoError(err)
		})
	}
}

func (s *TypesTestSuite) TestMsgEditSophon_ValidateBasic() {
	tests := []struct {
		name    string
		msg     *types.MsgEditSophon
		wantErr error
	}{
		{
			name: "Update admin address",
			msg: &types.MsgEditSophon{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewAdminAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				SophonPublicKey: types.DoNotModifyField,
				NewAddress:      types.DoNotModifyField,
				NewMemo:         types.DoNotModifyField,
				NewPublicKey:    types.DoNotModifyField,
			},
		},
		{
			name: "Update admin address and memo",
			msg: &types.MsgEditSophon{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewAdminAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				SophonPublicKey: types.DoNotModifyField,
				NewAddress:      types.DoNotModifyField,
				NewMemo:         "This is a sweet sophon",
				NewPublicKey:    types.DoNotModifyField,
			},
		},
		{
			name: "No updates",
			msg: &types.MsgEditSophon{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewAdminAddress: types.DoNotModifyField,
				SophonPublicKey: types.DoNotModifyField,
				NewAddress:      types.DoNotModifyField,
				NewMemo:         types.DoNotModifyField,
				NewPublicKey:    types.DoNotModifyField,
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "Invalid owner address",
			msg: &types.MsgEditSophon{
				OwnerAddress:    "invalid",
				NewAdminAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				SophonPublicKey: types.DoNotModifyField,
				NewAddress:      types.DoNotModifyField,
				NewMemo:         types.DoNotModifyField,
				NewPublicKey:    types.DoNotModifyField,
			},
			wantErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "Invalid admin address",
			msg: &types.MsgEditSophon{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewAdminAddress: "invalid",
				SophonPublicKey: types.DoNotModifyField,
				NewAddress:      types.DoNotModifyField,
				NewMemo:         types.DoNotModifyField,
				NewPublicKey:    types.DoNotModifyField,
			},
			wantErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "Invalid address",
			msg: &types.MsgEditSophon{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewAdminAddress: types.DoNotModifyField,
				SophonPublicKey: types.DoNotModifyField,
				NewAddress:      "invalid",
				NewMemo:         types.DoNotModifyField,
				NewPublicKey:    types.DoNotModifyField,
			},
			wantErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "Invalid memo",
			msg: &types.MsgEditSophon{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewAdminAddress: types.DoNotModifyField,
				SophonPublicKey: types.DoNotModifyField,
				NewAddress:      types.DoNotModifyField,
				NewMemo:         strings.Repeat("a", 3001),
				NewPublicKey:    types.DoNotModifyField,
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "Empty public key",
			msg: &types.MsgEditSophon{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewAdminAddress: types.DoNotModifyField,
				SophonPublicKey: types.DoNotModifyField,
				NewAddress:      types.DoNotModifyField,
				NewMemo:         types.DoNotModifyField,
				NewPublicKey:    "",
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "Invalid public key",
			msg: &types.MsgEditSophon{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewAdminAddress: types.DoNotModifyField,
				SophonPublicKey: types.DoNotModifyField,
				NewAddress:      types.DoNotModifyField,
				NewMemo:         types.DoNotModifyField,
				NewPublicKey:    "not valid hex",
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := test.msg.ValidateBasic()

			if test.wantErr != nil {
				s.Require().ErrorIs(err, test.wantErr)
				return
			}

			s.Require().NoError(err)
		})
	}
}

func (s *TypesTestSuite) TestMsgTransferOwnership_ValidateBasic() {
	pubKeyHex := "041b84c5567b126440995d3ed5aaba0565d71e1834604819ff9c17f5e9d5dd078f70beaf8f588b541507fed6a642c5ab42dfdf8120a7f639de5122d47a69a8e8d1"

	tests := []struct {
		name    string
		msg     *types.MsgTransferOwnership
		wantErr error
	}{
		{
			name: "valid",
			msg: &types.MsgTransferOwnership{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewOwnerAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				SophonPublicKey: pubKeyHex,
			},
		},
		{
			name: "new owner address is the same as the current owner address",
			msg: &types.MsgTransferOwnership{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewOwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				SophonPublicKey: pubKeyHex,
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid owner address",
			msg: &types.MsgTransferOwnership{
				OwnerAddress:    "invalid",
				NewOwnerAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				SophonPublicKey: pubKeyHex,
			},
			wantErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid new owner address",
			msg: &types.MsgTransferOwnership{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewOwnerAddress: "invalid",
				SophonPublicKey: pubKeyHex,
			},
			wantErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid sophon public key",
			msg: &types.MsgTransferOwnership{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewOwnerAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				SophonPublicKey: "not valid hex",
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "empty sophon public key",
			msg: &types.MsgTransferOwnership{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewOwnerAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				SophonPublicKey: "",
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := test.msg.ValidateBasic()

			if test.wantErr != nil {
				s.Require().ErrorIs(err, test.wantErr)
				return
			}

			s.Require().NoError(err)
		})
	}
}

func (s *TypesTestSuite) TestMsgAcceptOwnership_ValidateBasic() {
	pubKeyHex := "041b84c5567b126440995d3ed5aaba0565d71e1834604819ff9c17f5e9d5dd078f70beaf8f588b541507fed6a642c5ab42dfdf8120a7f639de5122d47a69a8e8d1"

	tests := []struct {
		name    string
		msg     *types.MsgAcceptOwnership
		wantErr error
	}{
		{
			name: "valid",
			msg: &types.MsgAcceptOwnership{
				NewOwnerAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				SophonPublicKey: pubKeyHex,
			},
		},
		{
			name: "invalid new owner address",
			msg: &types.MsgAcceptOwnership{
				NewOwnerAddress: "invalid",
				SophonPublicKey: pubKeyHex,
			},
			wantErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid sophon public key",
			msg: &types.MsgAcceptOwnership{
				NewOwnerAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				SophonPublicKey: "not valid hex",
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "empty sophon public key",
			msg: &types.MsgAcceptOwnership{
				NewOwnerAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				SophonPublicKey: "",
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := test.msg.ValidateBasic()

			if test.wantErr != nil {
				s.Require().ErrorIs(err, test.wantErr)
				return
			}

			s.Require().NoError(err)
		})
	}
}

func (s *TypesTestSuite) TestMsgCancelOwnershipTransfer_ValidateBasic() {
	pubKeyHex := "041b84c5567b126440995d3ed5aaba0565d71e1834604819ff9c17f5e9d5dd078f70beaf8f588b541507fed6a642c5ab42dfdf8120a7f639de5122d47a69a8e8d1"

	tests := []struct {
		name    string
		msg     *types.MsgCancelOwnershipTransfer
		wantErr error
	}{
		{
			name: "valid",
			msg: &types.MsgCancelOwnershipTransfer{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				SophonPublicKey: pubKeyHex,
			},
		},
		{
			name: "invalid owner address",
			msg: &types.MsgCancelOwnershipTransfer{
				OwnerAddress:    "invalid",
				SophonPublicKey: pubKeyHex,
			},
			wantErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid sophon public key",
			msg: &types.MsgCancelOwnershipTransfer{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				SophonPublicKey: "not valid hex",
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "empty sophon public key",
			msg: &types.MsgCancelOwnershipTransfer{
				OwnerAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				SophonPublicKey: "",
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := test.msg.ValidateBasic()

			if test.wantErr != nil {
				s.Require().ErrorIs(err, test.wantErr)
				return
			}

			s.Require().NoError(err)
		})
	}
}
