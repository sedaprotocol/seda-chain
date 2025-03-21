package keeper_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto/ed25519"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdktestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/std"
	sdkintegration "github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	sdkstakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	sdkstakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm/v2"

	"github.com/sedaprotocol/seda-chain/app"
	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/testutil"
	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	batchingkeeper "github.com/sedaprotocol/seda-chain/x/batching/keeper"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	dataproxykeeper "github.com/sedaprotocol/seda-chain/x/data-proxy/keeper"
	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"
	pubkeykeeper "github.com/sedaprotocol/seda-chain/x/pubkey/keeper"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
	"github.com/sedaprotocol/seda-chain/x/staking"
	stakingkeeper "github.com/sedaprotocol/seda-chain/x/staking/keeper"
	"github.com/sedaprotocol/seda-chain/x/tally"
	"github.com/sedaprotocol/seda-chain/x/tally/keeper"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
	wasmstorage "github.com/sedaprotocol/seda-chain/x/wasm-storage"
	wasmstoragekeeper "github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

const (
	bech32Prefix = "seda"
	bondDenom    = "aseda"
)

type fixture struct {
	*testutil.IntegationApp
	cdc               codec.Codec
	txConfig          client.TxConfig
	chainID           string
	coreContractAddr  sdk.AccAddress
	deployer          sdk.AccAddress
	accountKeeper     authkeeper.AccountKeeper
	bankKeeper        bankkeeper.Keeper
	stakingKeeper     stakingkeeper.Keeper
	contractKeeper    wasmkeeper.PermissionedKeeper
	wasmKeeper        wasmkeeper.Keeper
	wasmStorageKeeper wasmstoragekeeper.Keeper
	tallyKeeper       keeper.Keeper
	tallyMsgServer    types.MsgServer
	batchingKeeper    batchingkeeper.Keeper
	dataProxyKeeper   *dataproxykeeper.Keeper
	wasmViewKeeper    wasmtypes.ViewKeeper
	logBuf            *bytes.Buffer
}

func initFixture(t testing.TB) *fixture {
	t.Helper()

	tempDir := t.TempDir()

	chainID := "integration-app"
	tallyvm.TallyMaxBytes = 1024

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, sdkstakingtypes.StoreKey, wasmstoragetypes.StoreKey,
		wasmtypes.StoreKey, pubkeytypes.StoreKey, batchingtypes.StoreKey, types.StoreKey,
		dataproxytypes.StoreKey,
	)

	mb := module.NewBasicManager(auth.AppModuleBasic{}, bank.AppModuleBasic{}, wasmstorage.AppModuleBasic{}, wasm.AppModuleBasic{})

	interfaceRegistry := sdktestutil.CodecOptions{
		AccAddressPrefix: params.Bech32PrefixAccAddr,
		ValAddressPrefix: params.Bech32PrefixValAddr,
	}.NewInterfaceRegistry()
	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	aminoCodec := codec.NewLegacyAmino()
	encCfg := moduletestutil.TestEncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             protoCodec,
		TxConfig:          tx.NewTxConfig(protoCodec, tx.DefaultSignModes),
		Amino:             aminoCodec,
	}
	cdc := encCfg.Codec
	txConfig := encCfg.TxConfig
	std.RegisterLegacyAminoCodec(encCfg.Amino)
	std.RegisterInterfaces(encCfg.InterfaceRegistry)
	mb.RegisterLegacyAminoCodec(encCfg.Amino)
	mb.RegisterInterfaces(encCfg.InterfaceRegistry)

	buf := &bytes.Buffer{}
	logger := log.NewLogger(buf, log.LevelOption(zerolog.DebugLevel))

	cms := sdkintegration.CreateMultiStore(keys, logger)

	ctx := sdk.NewContext(cms, cmtproto.Header{Time: time.Now().UTC()}, true, logger)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	maccPerms := map[string][]string{
		authtypes.FeeCollectorName:        nil,
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

	sdkStakingKeeper := sdkstakingkeeper.NewKeeper(cdc, runtime.NewKVStoreService(keys[sdkstakingtypes.StoreKey]), accountKeeper, bankKeeper, authority.String(), addresscodec.NewBech32Codec(params.Bech32PrefixValAddr), addresscodec.NewBech32Codec(params.Bech32PrefixConsAddr))
	stakingKeeper := stakingkeeper.NewKeeper(sdkStakingKeeper, addresscodec.NewBech32Codec(params.Bech32PrefixValAddr))

	stakingParams := sdkstakingtypes.DefaultParams()
	stakingParams.BondDenom = bondDenom
	err := stakingKeeper.SetParams(ctx, stakingParams)
	require.NoError(t, err)

	// x/wasm
	router := baseapp.NewMsgServiceRouter()
	wasmKeeper := wasmkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		nil, nil, nil, nil,
		nil, nil, router, nil,
		tempDir,
		wasmtypes.DefaultWasmConfig(),
		app.GetWasmCapabilities(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		[]wasmkeeper.Option{}...,
	)
	require.NoError(t, wasmKeeper.SetParams(ctx, wasmtypes.DefaultParams()))

	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(&wasmKeeper)

	wasmStorageKeeper := wasmstoragekeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[wasmstoragetypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authtypes.FeeCollectorName,
		encCfg.TxConfig.TxDecoder(),
		bankKeeper,
		stakingKeeper,
		contractKeeper,
	)

	slashingKeeper := slashingkeeper.NewKeeper(
		cdc,
		nil,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		stakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	pubKeyKeeper := pubkeykeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[pubkeytypes.StoreKey]),
		stakingKeeper,
		slashingKeeper,
		addresscodec.NewBech32Codec(params.Bech32PrefixValAddr),
		authtypes.NewModuleAddress("gov").String(),
	)
	stakingKeeper.SetPubKeyKeeper(pubKeyKeeper)

	dataProxyKeeper := dataproxykeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[dataproxytypes.StoreKey]),
		authtypes.NewModuleAddress("gov").String(),
	)

	batchingKeeper := batchingkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[batchingtypes.StoreKey]),
		stakingKeeper,
		slashingKeeper,
		wasmStorageKeeper,
		pubKeyKeeper,
		contractKeeper,
		wasmKeeper,
		addresscodec.NewBech32Codec(params.Bech32PrefixValAddr),
	)

	tallyKeeper := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[types.StoreKey]),
		wasmStorageKeeper,
		batchingKeeper,
		dataProxyKeeper,
		contractKeeper,
		wasmKeeper,
		authority.String(),
	)

	tallyMsgServer := keeper.NewMsgServerImpl(tallyKeeper)

	authModule := auth.NewAppModule(cdc, accountKeeper, app.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, pubKeyKeeper)
	wasmStorageModule := wasmstorage.NewAppModule(cdc, *wasmStorageKeeper)
	tallyModule := tally.NewAppModule(cdc, tallyKeeper)

	integrationApp := testutil.NewIntegrationApp(ctx, logger, keys, cdc, router, map[string]appmodule.AppModule{
		authtypes.ModuleName:        authModule,
		banktypes.ModuleName:        bankModule,
		sdkstakingtypes.ModuleName:  stakingModule,
		wasmstoragetypes.ModuleName: wasmStorageModule,
		types.ModuleName:            tallyModule,
	})

	// TODO: Check why IntegrationApp setup fails to initialize params.
	bankKeeper.SetSendEnabled(ctx, "aseda", true)

	err = tallyKeeper.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	// Upload, instantiate, and configure the Core Contract.
	deployer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	int1e21, ok := math.NewIntFromString("1000000000000000000000")
	require.True(t, ok)
	err = bankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(bondDenom, int1e21)))
	require.NoError(t, err)
	err = bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, deployer, sdk.NewCoins(sdk.NewCoin(bondDenom, int1e21)))
	require.NoError(t, err)

	codeID, _, err := contractKeeper.Create(ctx, deployer, testwasms.CoreContractWasm(), nil)
	require.NoError(t, err)
	require.Equal(t, uint64(1), codeID)

	initMsg := struct {
		Token   string         `json:"token"`
		Owner   sdk.AccAddress `json:"owner"`
		ChainID string         `json:"chain_id"`
	}{
		Token:   "aseda",
		Owner:   deployer,
		ChainID: chainID,
	}
	initMsgBz, err := json.Marshal(initMsg)
	require.NoError(t, err)

	coreContractAddr, _, err := contractKeeper.Instantiate(ctx, codeID, deployer, nil, initMsgBz, "Core Contract", sdk.NewCoins())
	require.NoError(t, err)
	require.NotEmpty(t, coreContractAddr)

	err = wasmStorageKeeper.CoreContractRegistry.Set(ctx, coreContractAddr.String())
	require.NoError(t, err)

	_, err = contractKeeper.Execute(
		ctx,
		coreContractAddr,
		deployer,
		[]byte(setStakingConfigMsg),
		sdk.NewCoins(),
	)
	require.NoError(t, err)

	return &fixture{
		IntegationApp:     integrationApp,
		chainID:           chainID,
		deployer:          deployer,
		cdc:               cdc,
		txConfig:          txConfig,
		coreContractAddr:  coreContractAddr,
		accountKeeper:     accountKeeper,
		bankKeeper:        bankKeeper,
		stakingKeeper:     *stakingKeeper,
		contractKeeper:    *contractKeeper,
		wasmKeeper:        wasmKeeper,
		wasmStorageKeeper: *wasmStorageKeeper,
		tallyKeeper:       tallyKeeper,
		tallyMsgServer:    tallyMsgServer,
		batchingKeeper:    batchingKeeper,
		dataProxyKeeper:   dataProxyKeeper,
		wasmViewKeeper:    wasmKeeper,
		logBuf:            buf,
	}
}

func (f *fixture) SetDataProxyConfig(proxyPubKey, payoutAddr string, proxyFee sdk.Coin) error {
	pkBytes, err := hex.DecodeString(proxyPubKey)
	if err != nil {
		return err
	}
	err = f.dataProxyKeeper.SetDataProxyConfig(f.Context(), pkBytes,
		dataproxytypes.ProxyConfig{
			PayoutAddress: payoutAddr,
			Fee:           &proxyFee,
		},
	)
	return err
}

var setStakingConfigMsg = `{
	"set_staking_config": {
	  "minimum_stake_to_register": "1",
	  "minimum_stake_for_committee_eligibility": "1",
	  "allowlist_enabled": true
	}
  }`
