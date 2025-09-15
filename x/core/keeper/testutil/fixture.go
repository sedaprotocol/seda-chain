package testutil

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdkwasm "github.com/CosmWasm/wasmd/x/wasm"
	sdkwasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
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

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm/v3"

	"github.com/sedaprotocol/seda-chain/app"
	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/testutil"
	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	batchingkeeper "github.com/sedaprotocol/seda-chain/x/batching/keeper"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	"github.com/sedaprotocol/seda-chain/x/core"
	"github.com/sedaprotocol/seda-chain/x/core/keeper"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	dataproxykeeper "github.com/sedaprotocol/seda-chain/x/data-proxy/keeper"
	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"
	pubkeykeeper "github.com/sedaprotocol/seda-chain/x/pubkey/keeper"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
	"github.com/sedaprotocol/seda-chain/x/staking"
	stakingkeeper "github.com/sedaprotocol/seda-chain/x/staking/keeper"
	"github.com/sedaprotocol/seda-chain/x/wasm"
	wasmstorage "github.com/sedaprotocol/seda-chain/x/wasm-storage"
	wasmstoragekeeper "github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

const (
	Bech32Prefix = "seda"
	BondDenom    = "aseda"
)

type Fixture struct {
	tb testing.TB
	*testutil.IntegationApp
	Codec             codec.Codec
	TxConfig          client.TxConfig
	ChainID           string
	CoreContractAddr  sdk.AccAddress
	Stakers           []Staker
	AccountKeeper     authkeeper.AccountKeeper
	BankKeeper        bankkeeper.Keeper
	StakingKeeper     stakingkeeper.Keeper
	ContractKeeper    sdkwasmkeeper.PermissionedKeeper
	WasmKeeper        sdkwasmkeeper.Keeper
	WasmStorageKeeper wasmstoragekeeper.Keeper
	CoreKeeper        keeper.Keeper
	CoreMsgServer     types.MsgServer
	CoreQuerier       types.QueryServer
	BatchingKeeper    batchingkeeper.Keeper
	DataProxyKeeper   *dataproxykeeper.Keeper
	WasmViewKeeper    wasmtypes.ViewKeeper
	LogBuf            *bytes.Buffer
	Router            *baseapp.MsgServiceRouter
	Creator           TestAccount
	Deployer          TestAccount
	TestAccounts      map[string]TestAccount
}

func InitFixture(tb testing.TB) *Fixture {
	tb.Helper()

	tempDir := tb.TempDir()

	chainID := "integration-app"
	tallyvm.TallyMaxBytes = 1024

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, sdkstakingtypes.StoreKey, wasmstoragetypes.StoreKey,
		wasmtypes.StoreKey, pubkeytypes.StoreKey, batchingtypes.StoreKey, dataproxytypes.StoreKey,
		types.StoreKey,
	)

	mb := module.NewBasicManager(
		auth.AppModuleBasic{}, bank.AppModuleBasic{}, wasmstorage.AppModuleBasic{},
		sdkwasm.AppModuleBasic{}, core.AppModuleBasic{},
	)

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
	std.RegisterLegacyAminoCodec(encCfg.Amino)
	std.RegisterInterfaces(encCfg.InterfaceRegistry)
	mb.RegisterLegacyAminoCodec(encCfg.Amino)
	mb.RegisterInterfaces(encCfg.InterfaceRegistry)

	buf := &bytes.Buffer{}
	logger := log.NewLogger(buf, log.LevelOption(zerolog.DebugLevel))

	cms := sdkintegration.CreateMultiStore(keys, logger)

	ctx := sdk.NewContext(cms, cmtproto.Header{Time: time.Now().UTC(), ChainID: chainID}, true, logger)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	creator := TestAccount{
		name:       "creator",
		addr:       authority,
		signingKey: secp256k1.GenPrivKey(),
		fixture:    nil,
		Sequence:   0,
	}

	maccPerms := map[string][]string{
		authtypes.FeeCollectorName:        nil,
		minttypes.ModuleName:              {authtypes.Minter},
		sdkstakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		sdkstakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		wasmtypes.ModuleName:              {authtypes.Burner},
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

	sdkStakingKeeper := sdkstakingkeeper.NewKeeper(cdc, runtime.NewKVStoreService(keys[sdkstakingtypes.StoreKey]), accountKeeper, bankKeeper, authority.String(), addresscodec.NewBech32Codec(params.Bech32PrefixValAddr), addresscodec.NewBech32Codec(params.Bech32PrefixConsAddr))
	stakingKeeper := stakingkeeper.NewKeeper(sdkStakingKeeper, addresscodec.NewBech32Codec(params.Bech32PrefixValAddr))

	stakingParams := sdkstakingtypes.DefaultParams()
	stakingParams.BondDenom = BondDenom
	err := stakingKeeper.SetParams(ctx, stakingParams)
	require.NoError(tb, err)

	// x/wasm
	wasmStoreService := runtime.NewKVStoreService(keys[wasmtypes.StoreKey])
	router := baseapp.NewMsgServiceRouter()
	sdkWasmKeeper := sdkwasmkeeper.NewKeeper(
		cdc,
		wasmStoreService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		nil, nil, nil, nil,
		nil, nil, router, nil,
		tempDir,
		wasmtypes.DefaultWasmConfig(),
		app.GetWasmCapabilities(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		[]sdkwasmkeeper.Option{}...,
	)

	wasmKeeper := wasm.NewKeeper(
		&sdkWasmKeeper,
		stakingKeeper,
		cdc,
		router,
		nil, // queryRouter
		wasmStoreService,
	)

	contractKeeper := sdkwasmkeeper.NewDefaultPermissionKeeper(wasmKeeper)

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

	wasmKeeper.SetWasmStorageKeeper(wasmStorageKeeper)
	require.NoError(tb, wasmKeeper.SetParams(ctx, wasmtypes.DefaultParams()))

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
		bankKeeper,
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

	coreKeeper := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[types.StoreKey]),
		wasmStorageKeeper,
		batchingKeeper,
		dataProxyKeeper,
		stakingKeeper,
		bankKeeper,
		contractKeeper,
		wasmKeeper,
		authority.String(),
		authtypes.FeeCollectorName,
		encCfg.TxConfig.TxDecoder(),
	)

	genesis := types.DefaultGenesisState()
	genesis.Owner = authority.String()
	coreKeeper.InitGenesis(ctx, *genesis)

	coreMsgServer := keeper.NewMsgServerImpl(coreKeeper)

	coreQuerier := keeper.Querier{Keeper: coreKeeper}

	authModule := auth.NewAppModule(cdc, accountKeeper, app.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, pubKeyKeeper)
	wasmStorageModule := wasmstorage.NewAppModule(cdc, *wasmStorageKeeper)
	wasmModule := wasm.NewAppModule(cdc, wasmKeeper, stakingKeeper, accountKeeper, bankKeeper, router, nil)
	coreModule := core.NewAppModule(cdc, coreKeeper)

	integrationApp := testutil.NewIntegrationApp(
		ctx, logger, keys, cdc, router,
		map[string]appmodule.AppModule{
			authtypes.ModuleName:        authModule,
			banktypes.ModuleName:        bankModule,
			sdkstakingtypes.ModuleName:  stakingModule,
			wasmstoragetypes.ModuleName: wasmStorageModule,
			wasmtypes.ModuleName:        wasmModule,
			types.ModuleName:            coreModule,
		},
	)
	wasmKeeper.SetRouter(router)

	// TODO: Check why IntegrationApp setup fails to initialize params.
	err = pubKeyKeeper.SetProvingScheme(
		ctx,
		pubkeytypes.ProvingScheme{
			Index:            0, // SEDA Key Index for Secp256k1
			IsActivated:      true,
			ActivationHeight: ctx.BlockHeight(),
		},
	)
	require.NoError(tb, err)

	// TODO: Check why IntegrationApp setup fails to initialize params.
	bankKeeper.SetSendEnabled(ctx, "aseda", true)

	err = coreKeeper.SetParams(ctx, types.DefaultParams())
	require.NoError(tb, err)

	// Upload, instantiate, and configure the Core Contract.
	int1e21, ok := math.NewIntFromString("10000000000000000000000000")
	require.True(tb, ok)
	err = bankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(BondDenom, int1e21)))
	require.NoError(tb, err)
	err = bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, authority, sdk.NewCoins(sdk.NewCoin(BondDenom, int1e21)))
	require.NoError(tb, err)

	codeID, _, err := contractKeeper.Create(ctx, authority, testwasms.CoreContractWasm(), nil)
	require.NoError(tb, err)
	require.Equal(tb, uint64(1), codeID)

	initMsg := struct {
		Token   string         `json:"token"`
		Owner   sdk.AccAddress `json:"owner"`
		ChainID string         `json:"chain_id"`
	}{
		Token:   "aseda",
		Owner:   authority,
		ChainID: chainID,
	}
	initMsgBz, err := json.Marshal(initMsg)
	require.NoError(tb, err)

	coreContractAddr, _, err := contractKeeper.Instantiate(ctx, codeID, authority, nil, initMsgBz, "Core Contract", sdk.NewCoins())
	require.NoError(tb, err)
	require.NotEmpty(tb, coreContractAddr)

	err = wasmStorageKeeper.CoreContractRegistry.Set(ctx, coreContractAddr.String())
	require.NoError(tb, err)

	_, err = contractKeeper.Execute(
		ctx,
		coreContractAddr,
		authority,
		[]byte(setStakingConfigMsg),
		sdk.NewCoins(),
	)
	require.NoError(tb, err)

	f := Fixture{
		tb:                tb,
		IntegationApp:     integrationApp,
		ChainID:           chainID,
		Codec:             cdc,
		TxConfig:          encCfg.TxConfig,
		CoreContractAddr:  coreContractAddr,
		AccountKeeper:     accountKeeper,
		BankKeeper:        bankKeeper,
		StakingKeeper:     *stakingKeeper,
		ContractKeeper:    *contractKeeper,
		WasmKeeper:        *wasmKeeper.Keeper,
		WasmStorageKeeper: *wasmStorageKeeper,
		CoreKeeper:        coreKeeper,
		CoreMsgServer:     coreMsgServer,
		CoreQuerier:       coreQuerier,
		BatchingKeeper:    batchingKeeper,
		DataProxyKeeper:   dataProxyKeeper,
		WasmViewKeeper:    wasmKeeper,
		LogBuf:            buf,
		Router:            router,
		TestAccounts:      make(map[string]TestAccount),
		Creator:           creator,
	}
	f.Creator.fixture = &f
	f.Deployer.fixture = &f

	f.SetContextChainID(chainID)
	f.uploadOraclePrograms(tb)
	return &f
}

func (f *Fixture) GetTestAccount(name string) TestAccount {
	acc, exists := f.TestAccounts[name]
	require.True(f.tb, exists, "test account %s not found", name)
	return acc
}

func (f *Fixture) CreateTestAccount(name string, initialBalanceSeda int64) TestAccount {
	addrPrivKey := ed25519.GenPrivKey()
	addr := sdk.AccAddress(addrPrivKey.PubKey().Address())

	acc := TestAccount{
		name:       name,
		addr:       addr,
		signingKey: secp256k1.GenPrivKey(),
		fixture:    f,
		Sequence:   0,
	}
	bigAmountSeda := math.NewInt(initialBalanceSeda)
	bigAmount := bigAmountSeda.Mul(math.NewInt(1_000_000_000_000_000_000))
	f.initAccountWithCoins(f.tb, acc.AccAddress(), sdk.NewCoins(sdk.NewCoin(BondDenom, bigAmount)))

	f.TestAccounts[name] = acc
	return acc
}

func (f *Fixture) SetDataProxyConfig(proxyPubKey, payoutAddr string, proxyFee sdk.Coin) error {
	pkBytes, err := hex.DecodeString(proxyPubKey)
	if err != nil {
		return err
	}
	err = f.DataProxyKeeper.SetDataProxyConfig(f.Context(), pkBytes,
		dataproxytypes.ProxyConfig{
			PayoutAddress: payoutAddr,
			Fee:           &proxyFee,
		},
	)
	return err
}

var setStakingConfigMsg = `{
	"set_staking_config": {
	  "minimum_stake": "1",
	  "allowlist_enabled": true
	}
  }`
