package staking

import (
	"github.com/CosmWasm/wasmd/x/wasm/exported"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	_ module.AppModule = AppModule{}
)

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements an application module for the staking module.
type AppModule struct {
	staking.AppModule
	stakingKeeper *stakingkeeper.Keeper
	accountKeeper AccountKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec,
	keeper *keeper.Keeper,
	ak AccountKeeper,
	bk stakingtypes.BankKeeper,
	ls exported.Subspace,
) AppModule {
	am := staking.NewAppModule(cdc, keeper, ak, bk, ls)
	return AppModule{
		AppModule:     am,
		stakingKeeper: keeper,
		accountKeeper: ak,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	stakingtypes.RegisterMsgServer(cfg.MsgServer(), NewMsgServerImpl(am.stakingKeeper, am.accountKeeper))

	querier := stakingkeeper.Querier{Keeper: am.stakingKeeper}
	stakingtypes.RegisterQueryServer(cfg.QueryServer(), querier)
}
