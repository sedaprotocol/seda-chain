package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(_ *codec.LegacyAmino) {
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAcceptOwnership{},
		&MsgTransferOwnership{},
		&MsgAddToAllowlist{},
		&MsgRemoveFromAllowlist{},
		&MsgPause{},
		&MsgUnpause{},
		&MsgStake{},
		&MsgLegacyStake{},
		&MsgUnstake{},
		&MsgLegacyUnstake{},
		&MsgWithdraw{},
		&MsgLegacyWithdraw{},
		&MsgPostDataRequest{},
		&MsgCommit{},
		&MsgLegacyCommit{},
		&MsgReveal{},
		&MsgLegacyReveal{},
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
