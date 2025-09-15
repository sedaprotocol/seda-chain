package wasm

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/exported"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/simulation"
	"github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var _ module.AppModule = AppModule{}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule wraps the default AppModule to provide a custom MsgServer that
// redirects Core Contract messages to x/core.
type AppModule struct {
	wasm.AppModule

	keeper *Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec,
	keeper *Keeper,
	validatorSetSource keeper.ValidatorSetSource,
	ak types.AccountKeeper,
	bk simulation.BankKeeper,
	router *baseapp.MsgServiceRouter,
	ss exported.Subspace,
) AppModule {
	return AppModule{
		AppModule: wasm.NewAppModule(cdc, keeper.Keeper, validatorSetSource, ak, bk, router, ss),
		keeper:    keeper,
	}
}

// RegisterServices overrides the default RegisterServices method to register
// the custom MsgServer that redirects Core Contract messages to x/core.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	sdkMsgServer := keeper.NewMsgServerImpl(am.keeper.Keeper)
	types.RegisterMsgServer(cfg.MsgServer(), NewMsgServerImpl(sdkMsgServer, am.keeper))

	// TODO add querier shim
	types.RegisterQueryServer(cfg.QueryServer(), Querier(am.keeper))
}
