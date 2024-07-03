package keeper_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/cometbft/cometbft/crypto/ed25519"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
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
	"github.com/sedaprotocol/seda-chain/x/staking"
	stakingkeeper "github.com/sedaprotocol/seda-chain/x/staking/keeper"
	wasmstorage "github.com/sedaprotocol/seda-chain/x/wasm-storage"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper/testdata"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

var wasmCapabilities = []string{
	"iterator",
	"staking",
	"stargate",
	"cosmwasm_1_1",
	"cosmwasm_1_2",
	"cosmwasm_1_3",
	"cosmwasm_1_4",
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
	wasmStorageKeeper keeper.Keeper
	mockViewKeeper    *testutil.MockViewKeeper
	logBuf            *bytes.Buffer
}

func initFixture(tb testing.TB) *fixture {
	tb.Helper()

	tempDir := tb.TempDir()

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, sdkstakingtypes.StoreKey, types.StoreKey, wasmtypes.StoreKey,
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
		types.ModuleName:                  {authtypes.Burner},
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

	sdkstakingKeeper := sdkstakingkeeper.NewKeeper(cdc, runtime.NewKVStoreService(keys[sdkstakingtypes.StoreKey]), accountKeeper, bankKeeper, authority.String(), addresscodec.NewBech32Codec(params.Bech32PrefixValAddr), addresscodec.NewBech32Codec(params.Bech32PrefixConsAddr))
	stakingKeeper := stakingkeeper.NewKeeper(sdkstakingKeeper)

	stakingParams := sdkstakingtypes.DefaultParams()
	stakingParams.BondDenom = bondDenom
	err := stakingKeeper.SetParams(ctx, stakingParams)
	require.NoError(tb, err)

	// x/wasm
	wasmKeeper := wasmkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[types.StoreKey]),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		tempDir,
		wasmtypes.DefaultWasmConfig(), //wasmConfig,
		strings.Join(wasmCapabilities, ","),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		[]wasmkeeper.Option{}...,
	)
	require.NoError(tb, wasmKeeper.SetParams(ctx, wasmtypes.DefaultParams()))

	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(&wasmKeeper)

	// x/wasm-storage
	ctrl := gomock.NewController(tb)
	viewKeeper := testutil.NewMockViewKeeper(ctrl)

	wasmStorageKeeper := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[types.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		accountKeeper,
		bankKeeper,
		contractKeeper,
		viewKeeper,
	)

	authModule := auth.NewAppModule(cdc, accountKeeper, app.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, nil)
	wasmStorageModule := wasmstorage.NewAppModule(cdc, *wasmStorageKeeper)

	// Upload and instantiate the SEDA contract.
	creator := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	codeID, _, err := contractKeeper.Create(ctx, creator, testdata.SedaContractWasm(), nil)
	require.NoError(tb, err)
	require.Equal(tb, uint64(1), codeID)

	initMsg := struct {
		Token   string         `json:"token"`
		Owner   sdk.AccAddress `json:"owner"`
		ChainID string         `json:"chain_id"`
	}{
		Token:   "aseda",
		Owner:   sdk.MustAccAddressFromBech32("seda1xd04svzj6zj93g4eknhp6aq2yyptagcc2zeetj"),
		ChainID: "integration-app",
	}
	initMsgBz, err := json.Marshal(initMsg)
	require.NoError(tb, err)

	contractAddr, _, err := contractKeeper.Instantiate(ctx, codeID, creator, nil, initMsgBz, "DR Contract", sdk.NewCoins())
	require.NoError(tb, err)
	require.NotEmpty(tb, contractAddr)

	err = wasmStorageKeeper.ProxyContractRegistry.Set(ctx, contractAddr.String())
	require.NoError(tb, err)

	// Store the sample tally that will be used.
	tallyWasm := types.NewWasm(testdata.SampleTallyWasm(), types.WasmTypeDataRequest, ctx.BlockTime(), ctx.BlockHeight(), 100)
	err = wasmStorageKeeper.DataRequestWasm.Set(ctx, tallyWasm.Hash, tallyWasm)
	require.NoError(tb, err)

	integrationApp := integration.NewIntegrationApp(ctx, logger, keys, cdc, map[string]appmodule.AppModule{
		authtypes.ModuleName:       authModule,
		banktypes.ModuleName:       bankModule,
		sdkstakingtypes.ModuleName: stakingModule,
		types.ModuleName:           wasmStorageModule,
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
		mockViewKeeper:    viewKeeper,
		logBuf:            buf,
	}
}
