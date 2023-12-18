package staking

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/staking"
	sdkcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	sdkkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	sdktypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/x/staking/types"
)

const (
	consensusVersion uint64 = 5
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic defines the basic application module used by the staking module.
type AppModuleBasic struct {
	// staking.AppModuleBasic

	cdc codec.Codec
	ak  types.AccountKeeper
}

// Name returns the staking module's name.
func (AppModuleBasic) Name() string {
	return sdktypes.ModuleName
}

// RegisterLegacyAminoCodec registers the staking module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	sdktypes.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
	err := registry.EnsureRegistered(&types.MsgCreateValidatorWithVRF{})
	if err != nil {
		fmt.Println("error")
	}
	sdktypes.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the staking
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(sdktypes.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the staking module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data sdktypes.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", sdktypes.ModuleName, err)
	}

	return staking.ValidateGenesis(&data)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the staking module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := sdktypes.RegisterQueryHandlerClient(context.Background(), mux, sdktypes.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the staking module.
func (amb AppModuleBasic) GetTxCmd() *cobra.Command {
	return sdkcli.NewTxCmd(amb.cdc.InterfaceRegistry().SigningContext().ValidatorAddressCodec(), amb.cdc.InterfaceRegistry().SigningContext().AddressCodec())
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements an application module for the staking module.
type AppModule struct {
	AppModuleBasic
	// staking.AppModule

	keeper           *sdkkeeper.Keeper
	accountKeeper    types.AccountKeeper
	bankKeeper       sdktypes.BankKeeper
	randomnessKeeper types.RandomnessKeeper
	// legacySubspace is used solely for migration of x/params managed parameters
	// legacySubspace exported.Subspace
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec,
	keeper *keeper.Keeper,
	ak types.AccountKeeper,
	bk sdktypes.BankKeeper,
	rk types.RandomnessKeeper,
	// ls exported.Subspace,
) AppModule {
	// am := staking.NewAppModule(cdc, keeper, ak, bk, ls)
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc, ak: ak},
		// AppModule:     am,
		keeper:           keeper,
		accountKeeper:    ak,
		bankKeeper:       bk,
		randomnessKeeper: rk,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), types.NewMsgServerImpl(am.keeper, am.accountKeeper, am.randomnessKeeper))

	sdktypes.RegisterMsgServer(cfg.MsgServer(), NewMsgServerImpl(am.keeper, am.accountKeeper))

	querier := sdkkeeper.Querier{Keeper: am.keeper}
	sdktypes.RegisterQueryServer(cfg.QueryServer(), querier)
}

// // RegisterServices registers module services.
// func (am AppModule) RegisterServices(cfg module.Configurator) {
// 	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
// 	querier := keeper.Querier{Keeper: am.keeper}
// 	types.RegisterQueryServer(cfg.QueryServer(), querier)

// 	m := keeper.NewMigrator(am.keeper, am.legacySubspace)
// 	if err := cfg.RegisterMigration(types.ModuleName, 1, m.Migrate1to2); err != nil {
// 		panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", types.ModuleName, err))
// 	}
// 	if err := cfg.RegisterMigration(types.ModuleName, 2, m.Migrate2to3); err != nil {
// 		panic(fmt.Sprintf("failed to migrate x/%s from version 2 to 3: %v", types.ModuleName, err))
// 	}
// 	if err := cfg.RegisterMigration(types.ModuleName, 3, m.Migrate3to4); err != nil {
// 		panic(fmt.Sprintf("failed to migrate x/%s from version 3 to 4: %v", types.ModuleName, err))
// 	}
// 	if err := cfg.RegisterMigration(types.ModuleName, 4, m.Migrate4to5); err != nil {
// 		panic(fmt.Sprintf("failed to migrate x/%s from version 4 to 5: %v", types.ModuleName, err))
// 	}
// }

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterInvariants registers the staking module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// InitGenesis performs genesis initialization for the staking module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState sdktypes.GenesisState

	cdc.MustUnmarshalJSON(data, &genesisState)

	return am.keeper.InitGenesis(ctx, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the staking
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(am.keeper.ExportGenesis(ctx))
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return consensusVersion }

// BeginBlock returns the begin blocker for the staking module.
func (am AppModule) BeginBlock(ctx context.Context) error {
	return am.keeper.BeginBlocker(ctx)
}

// EndBlock returns the end blocker for the staking module. It returns no validator
// updates.
func (am AppModule) EndBlock(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	return am.keeper.EndBlocker(ctx)
}
