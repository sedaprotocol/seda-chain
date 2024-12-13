package slashing

import (
	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
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
	registry cdctypes.InterfaceRegistry,
	pk PubKeyKeeper,
	valAddrCdc addresscodec.Codec,
) AppModule {
	baseAppModule := slashing.NewAppModule(cdc, keeper, ak, bk, sk, nil, registry)
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
	Registry              cdctypes.InterfaceRegistry
	ValidatorAddressCodec addresscodec.Codec
	LegacyAmino           *codec.LegacyAmino

	AccountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
	StakingKeeper types.StakingKeeper
	PubKeyKeeper  PubKeyKeeper
}

type ModuleOutputs struct {
	depinject.Out

	Keeper keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	k := keeper.NewKeeper(in.Cdc, in.LegacyAmino, in.StoreService, in.StakingKeeper, authority.String())
	m := NewAppModule(in.Cdc, k, in.AccountKeeper, in.BankKeeper, in.StakingKeeper, in.Registry, in.PubKeyKeeper, in.ValidatorAddressCodec)
	return ModuleOutputs{
		Keeper: k,
		Module: m,
	}
}
