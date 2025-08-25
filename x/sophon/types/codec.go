package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgRegisterSophon{}, "seda/sophon/MsgRegisterSophon")
	legacy.RegisterAminoMsg(cdc, &MsgEditSophon{}, "seda/sophon/MsgEditSophon")
	legacy.RegisterAminoMsg(cdc, &MsgTransferOwnership{}, "seda/sophon/MsgTransferOwnership")
	legacy.RegisterAminoMsg(cdc, &MsgAcceptOwnership{}, "seda/sophon/MsgAcceptOwnership")
	legacy.RegisterAminoMsg(cdc, &MsgCancelOwnershipTransfer{}, "seda/sophon/MsgCancelOwnershipTransfer")
	legacy.RegisterAminoMsg(cdc, &MsgAddUser{}, "seda/sophon/MsgAddUser")
	legacy.RegisterAminoMsg(cdc, &MsgTopUpUser{}, "seda/sophon/MsgTopUpUser")
	legacy.RegisterAminoMsg(cdc, &MsgSettleCredits{}, "seda/sophon/MsgSettleCredits")
	legacy.RegisterAminoMsg(cdc, &MsgExpireCredits{}, "seda/sophon/MsgExpireCredits")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitReports{}, "seda/sophon/MsgSubmitReports")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "seda/sophon/MsgUpdateParams")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterSophon{},
		&MsgEditSophon{},
		&MsgTransferOwnership{},
		&MsgAcceptOwnership{},
		&MsgCancelOwnershipTransfer{},
		&MsgAddUser{},
		&MsgTopUpUser{},
		&MsgSettleCredits{},
		&MsgExpireCredits{},
		&MsgSubmitReports{},
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
