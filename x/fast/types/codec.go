package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgRegisterFastClient{}, "seda/fast/MsgRegisterFastClient")
	legacy.RegisterAminoMsg(cdc, &MsgEditFastClient{}, "seda/fast/MsgEditFastClient")
	legacy.RegisterAminoMsg(cdc, &MsgTransferOwnership{}, "seda/fast/MsgTransferOwnership")
	legacy.RegisterAminoMsg(cdc, &MsgAcceptOwnership{}, "seda/fast/MsgAcceptOwnership")
	legacy.RegisterAminoMsg(cdc, &MsgCancelOwnershipTransfer{}, "seda/fast/MsgCancelOwnershipTransfer")
	legacy.RegisterAminoMsg(cdc, &MsgAddUser{}, "seda/fast/MsgAddUser")
	legacy.RegisterAminoMsg(cdc, &MsgTopUpUser{}, "seda/fast/MsgTopUpUser")
	legacy.RegisterAminoMsg(cdc, &MsgSettleCredits{}, "seda/fast/MsgSettleCredits")
	legacy.RegisterAminoMsg(cdc, &MsgExpireUserCredits{}, "seda/fast/MsgExpireUserCredits")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitReports{}, "seda/fast/MsgSubmitReports")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "seda/fast/MsgUpdateParams")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterFastClient{},
		&MsgEditFastClient{},
		&MsgTransferOwnership{},
		&MsgAcceptOwnership{},
		&MsgCancelOwnershipTransfer{},
		&MsgAddUser{},
		&MsgTopUpUser{},
		&MsgSettleCredits{},
		&MsgExpireUserCredits{},
		&MsgSubmitReports{},
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
