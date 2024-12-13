package staking

import (
	"context"
	"encoding/json"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	sdkstaking "github.com/cosmos/cosmos-sdk/x/staking"
	sdkkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	sdktypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/x/staking/client/cli"
	"github.com/sedaprotocol/seda-chain/x/staking/keeper"
	"github.com/sedaprotocol/seda-chain/x/staking/types"
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
	cdc codec.Codec
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
}

// DefaultGenesis returns default genesis state as raw bytes for the staking
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genesis := sdktypes.DefaultGenesisState()
	genesis.Params.BondDenom = params.DefaultBondDenom

	return cdc.MustMarshalJSON(sdktypes.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the staking module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var data sdktypes.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return errorsmod.Wrapf(err, "failed to unmarshal %s genesis state", sdktypes.ModuleName)
	}
	return sdkstaking.ValidateGenesis(&data)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the staking module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := sdktypes.RegisterQueryHandlerClient(context.Background(), mux, sdktypes.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the staking module.
func (amb AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd(
		amb.cdc.InterfaceRegistry().SigningContext().ValidatorAddressCodec(),
		amb.cdc.InterfaceRegistry().SigningContext().AddressCodec(),
	)
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements an application module for the staking module.
type AppModule struct {
	sdkstaking.AppModule
	AppModuleBasic

	keeper        *keeper.Keeper
	accountKeeper sdktypes.AccountKeeper
	bankKeeper    sdktypes.BankKeeper
	pubKeyKeeper  types.PubKeyKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec,
	keeper *keeper.Keeper,
	ak sdktypes.AccountKeeper,
	bk sdktypes.BankKeeper,
	pk types.PubKeyKeeper,
) AppModule {
	return AppModule{
		AppModule:      sdkstaking.NewAppModule(cdc, keeper.Keeper, ak, bk, nil),
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		accountKeeper:  ak,
		bankKeeper:     bk,
		pubKeyKeeper:   pk,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	sdkMsgServer := sdkkeeper.NewMsgServerImpl(am.keeper.Keeper)
	msgServer := keeper.NewMsgServerImpl(sdkMsgServer, am.keeper)

	sdktypes.RegisterMsgServer(cfg.MsgServer(), msgServer)
	types.RegisterMsgServer(cfg.MsgServer(), msgServer)

	querier := sdkkeeper.Querier{Keeper: am.keeper.Keeper}
	sdktypes.RegisterQueryServer(cfg.QueryServer(), querier)
}

// RegisterInvariants registers the staking module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// ----------------------------------------------------------------------------
// App Wiring Setup
// ----------------------------------------------------------------------------

var _ appmodule.AppModule = AppModule{}

func init() {
	appmodule.Register(&Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	StoreService          store.KVStoreService
	Cdc                   codec.Codec
	ValidatorAddressCodec addresscodec.Codec
	ConsensusAddressCodec addresscodec.Codec

	AccountKeeper sdktypes.AccountKeeper
	BankKeeper    sdktypes.BankKeeper
	PubKeyKeeper  types.PubKeyKeeper
}

type ModuleOutputs struct {
	depinject.Out

	Keeper *keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	sk := sdkkeeper.NewKeeper(
		in.Cdc,
		in.StoreService,
		in.AccountKeeper,
		in.BankKeeper,
		authority.String(),
		in.ValidatorAddressCodec,
		in.ConsensusAddressCodec,
	)
	k := keeper.NewKeeper(sk, in.ValidatorAddressCodec)
	m := NewAppModule(in.Cdc, k, in.AccountKeeper, in.BankKeeper, in.PubKeyKeeper)

	return ModuleOutputs{
		Keeper: k,
		Module: m,
	}
}
