package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgRegisterDataProxy{}, "sedachain/MsgRegisterDataProxy")
	legacy.RegisterAminoMsg(cdc, &MsgEditDataProxy{}, "sedachain/MsgEditDataProxy")
	legacy.RegisterAminoMsg(cdc, &MsgTransferAdmin{}, "sedachain/MsgTransferAdmin")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterDataProxy{},
		&MsgEditDataProxy{},
		&MsgTransferAdmin{},
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
