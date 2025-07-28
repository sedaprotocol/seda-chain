package wasm

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/exported"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/simulation"
	"github.com/CosmWasm/wasmd/x/wasm/types"
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
	keeper *keeper.Keeper,
	validatorSetSource keeper.ValidatorSetSource,
	ak types.AccountKeeper,
	bk simulation.BankKeeper,
	router *baseapp.MsgServiceRouter,
	ss exported.Subspace,
	wsk WasmStorageKeeper,
) AppModule {
	return AppModule{
		AppModule: wasm.NewAppModule(cdc, keeper, validatorSetSource, ak, bk, router, ss),
		keeper:    NewKeeper(keeper, wsk),
	}
}

// RegisterServices overrides the default RegisterServices method to register
// the custom MsgServer that redirects Core Contract messages to x/core.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	sdkMsgServer := keeper.NewMsgServerImpl(am.keeper.Keeper)
	types.RegisterMsgServer(cfg.MsgServer(), NewMsgServerImpl(sdkMsgServer, am.keeper))

	// types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	// types.RegisterQueryServer(cfg.QueryServer(), keeper.Querier(am.keeper))

	// m := keeper.NewMigrator(*am.keeper, am.legacySubspace)
	// err := cfg.RegisterMigration(types.ModuleName, 1, m.Migrate1to2)
	// if err != nil {
	// 	panic(err)
	// }
	// err = cfg.RegisterMigration(types.ModuleName, 2, m.Migrate2to3)
	// if err != nil {
	// 	panic(err)
	// }
	// err = cfg.RegisterMigration(types.ModuleName, 3, m.Migrate3to4)
	// if err != nil {
	// 	panic(err)
	// }
}
