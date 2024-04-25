package main

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	// ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	// ibcLightClient "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

func main() {}

func InitZone() {
	// Configure address prefixes.
	// cfg := sdk.GetConfig()
	// cfg.SetBech32PrefixForAccount(params.Bech32PrefixAccAddr, params.Bech32PrefixAccPub)
	// cfg.SetBech32PrefixForValidator(params.Bech32PrefixValAddr, params.Bech32PrefixValPub)
	// cfg.SetBech32PrefixForConsensusNode(params.Bech32PrefixConsAddr, params.Bech32PrefixConsPub)
	// cfg.Seal()
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// ibcclienttypes.RegisterInterfaces(registry)
	// ibcLightClient.RegisterInterfaces(registry)
	sdk.RegisterInterfaces(registry)
	txtypes.RegisterInterfaces(registry)
	cryptocodec.RegisterInterfaces(registry)
	banktypes.RegisterInterfaces(registry)

	// Prepare Proto codec.
	// interfaceRegistry, err := codectypes.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
	// 	ProtoFiles: proto.HybridResolver,
	// 	SigningOptions: signing.Options{
	// 		AddressCodec: address.Bech32Codec{
	// 			Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
	// 		},
	// 		ValidatorAddressCodec: address.Bech32Codec{
	// 			Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
	// 		},
	// 	},
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// std.RegisterInterfaces(interfaceRegistry)
	// app.ModuleBasics.RegisterInterfaces(interfaceRegistry)
	// filePlugin := &FilePlugin{
	// 	cdc: codec.NewProtoCodec(interfaceRegistry),
	// }
}
