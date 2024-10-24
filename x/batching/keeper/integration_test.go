package keeper_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdkintegration "github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	sdkstakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	sdkstakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app"
	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/integration"
	"github.com/sedaprotocol/seda-chain/x/batching"
	"github.com/sedaprotocol/seda-chain/x/batching/keeper"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
	"github.com/sedaprotocol/seda-chain/x/pubkey"
	pubkeykeeper "github.com/sedaprotocol/seda-chain/x/pubkey/keeper"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
	"github.com/sedaprotocol/seda-chain/x/staking"
	stakingkeeper "github.com/sedaprotocol/seda-chain/x/staking/keeper"
	"github.com/sedaprotocol/seda-chain/x/tally"
	tallykeeper "github.com/sedaprotocol/seda-chain/x/tally/keeper"
	tallytypes "github.com/sedaprotocol/seda-chain/x/tally/types"
	wasmstorage "github.com/sedaprotocol/seda-chain/x/wasm-storage"
	wasmstoragekeeper "github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper/testutil"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

var wasmCapabilities = []string{
	"iterator",
	"staking",
	"stargate",
	"cosmwasm_1_1",
	"cosmwasm_1_2",
	"cosmwasm_1_3",
	"cosmwasm_1_4",
	"cosmwasm_1_5",
	"cosmwasm_2_0",
	"cosmwasm_2_1",
}

const (
	bech32Prefix = "seda"
	bondDenom    = "aseda"
)

type fixture struct {
	*integration.IntegationApp
	cdc               codec.Codec
	accountKeeper     authkeeper.AccountKeeper
	bankKeeper        bankkeeper.Keeper
	stakingKeeper     stakingkeeper.Keeper
	contractKeeper    wasmkeeper.PermissionedKeeper
	wasmKeeper        wasmkeeper.Keeper
	wasmStorageKeeper wasmstoragekeeper.Keeper
	tallyKeeper       tallykeeper.Keeper
	pubKeyKeeper      pubkeykeeper.Keeper
	batchingKeeper    keeper.Keeper
	mockViewKeeper    *testutil.MockViewKeeper
	logBuf            *bytes.Buffer
}

func initFixture(tb testing.TB) *fixture {
	tb.Helper()

	tempDir := tb.TempDir()

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, sdkstakingtypes.StoreKey, wasmstoragetypes.StoreKey,
		wasmtypes.StoreKey, pubkeytypes.StoreKey, types.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}, wasmstorage.AppModuleBasic{}).Codec

	buf := &bytes.Buffer{}
	logger := log.NewLogger(buf, log.LevelOption(zerolog.DebugLevel))

	cms := sdkintegration.CreateMultiStore(keys, logger)

	ctx := sdk.NewContext(cms, cmtproto.Header{Time: time.Now().UTC()}, true, logger)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	maccPerms := map[string][]string{
		minttypes.ModuleName:              {authtypes.Minter},
		sdkstakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		sdkstakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		wasmtypes.ModuleName:              {authtypes.Burner},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(params.Bech32PrefixAccAddr),
		params.Bech32PrefixAccAddr,
		authority.String(),
	)

	blockedAddresses := map[string]bool{
		accountKeeper.GetAuthority(): false,
	}
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blockedAddresses,
		authority.String(),
		log.NewNopLogger(),
	)

	sdkStakingKeeper := sdkstakingkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[sdkstakingtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		authority.String(),
		addresscodec.NewBech32Codec(params.Bech32PrefixValAddr),
		addresscodec.NewBech32Codec(params.Bech32PrefixConsAddr),
	)
	stakingKeeper := stakingkeeper.NewKeeper(sdkStakingKeeper)

	stakingParams := sdkstakingtypes.DefaultParams()
	stakingParams.BondDenom = bondDenom
	err := stakingKeeper.SetParams(ctx, stakingParams)
	require.NoError(tb, err)

	// x/wasm
	wasmKeeper := wasmkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		nil, nil, nil, nil,
		nil, nil, nil, nil,
		tempDir,
		wasmtypes.DefaultWasmConfig(),
		wasmCapabilities,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		[]wasmkeeper.Option{}...,
	)
	require.NoError(tb, wasmKeeper.SetParams(ctx, wasmtypes.DefaultParams()))

	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(&wasmKeeper)

	ctrl := gomock.NewController(tb)
	viewKeeper := testutil.NewMockViewKeeper(ctrl)

	wasmStorageKeeper := wasmstoragekeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[wasmstoragetypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		accountKeeper,
		bankKeeper,
		contractKeeper,
		viewKeeper,
	)

	pubKeyKeeper := pubkeykeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[pubkeytypes.StoreKey]),
		stakingKeeper,
		addresscodec.NewBech32Codec(params.Bech32PrefixValAddr),
	)
	batchingKeeper := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[types.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		stakingKeeper,
		wasmStorageKeeper,
		pubKeyKeeper,
		contractKeeper,
		viewKeeper,
		addresscodec.NewBech32Codec(params.Bech32PrefixValAddr),
	)
	tallyKeeper := tallykeeper.NewKeeper(
		wasmStorageKeeper,
		batchingKeeper,
		contractKeeper,
		viewKeeper,
	)

	authModule := auth.NewAppModule(cdc, accountKeeper, app.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, nil)
	wasmStorageModule := wasmstorage.NewAppModule(cdc, *wasmStorageKeeper)
	tallyModule := tally.NewAppModule(tallyKeeper)
	pubKeyModule := pubkey.NewAppModule(cdc, *pubKeyKeeper)
	batchingModule := batching.NewAppModule(cdc, batchingKeeper)

	integrationApp := integration.NewIntegrationApp(ctx, logger, keys, cdc, map[string]appmodule.AppModule{
		authtypes.ModuleName:        authModule,
		banktypes.ModuleName:        bankModule,
		sdkstakingtypes.ModuleName:  stakingModule,
		wasmstoragetypes.ModuleName: wasmStorageModule,
		tallytypes.ModuleName:       tallyModule,
		pubkeytypes.ModuleName:      pubKeyModule,
		types.ModuleName:            batchingModule,
	})

	return &fixture{
		IntegationApp:     integrationApp,
		cdc:               cdc,
		accountKeeper:     accountKeeper,
		bankKeeper:        bankKeeper,
		stakingKeeper:     *stakingKeeper,
		contractKeeper:    *contractKeeper,
		wasmKeeper:        wasmKeeper,
		wasmStorageKeeper: *wasmStorageKeeper,
		tallyKeeper:       tallyKeeper,
		pubKeyKeeper:      *pubKeyKeeper,
		batchingKeeper:    batchingKeeper,
		mockViewKeeper:    viewKeeper,
		logBuf:            buf,
	}
}
