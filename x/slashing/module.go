package slashing

import (
	addresscodec "cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/slashing/exported"
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

var _ module.AppModule = AppModule{}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements an application module for the slashing module.
type AppModule struct {
	slashing.AppModule

	keeper                keeper.Keeper
	pubKeyKeeper          PubKeyKeeper
	validatorAddressCodec addresscodec.Codec
}

// NewAppModule creates a new AppModule object.
func NewAppModule(
	cdc codec.Codec,
	keeper keeper.Keeper,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
	ss exported.Subspace,
	registry cdctypes.InterfaceRegistry,
	pk PubKeyKeeper,
	valAddrCdc addresscodec.Codec,
) AppModule {
	baseAppModule := slashing.NewAppModule(cdc, keeper, ak, bk, sk, ss, registry)
	return AppModule{
		AppModule:             baseAppModule,
		keeper:                keeper,
		pubKeyKeeper:          pk,
		validatorAddressCodec: valAddrCdc,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	sdkMsgServer := keeper.NewMsgServerImpl(am.keeper)
	msgServer := NewMsgServerImpl(sdkMsgServer, am.pubKeyKeeper, am.validatorAddressCodec)
	types.RegisterMsgServer(cfg.MsgServer(), msgServer)
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}
