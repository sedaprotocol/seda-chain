package types

import (
	"cosmossdk.io/x/evidence/exported"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RegisterCodec(_ *codec.LegacyAmino) {
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&BatchDoubleSign{},
		&MsgUpdateParams{},
	)

	registry.RegisterImplementations(
		(*exported.Evidence)(nil),
		&BatchDoubleSign{},
	)
}
