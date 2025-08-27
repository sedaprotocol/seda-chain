package keeper_test

import (
	"bytes"
	"crypto/ecdsa"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
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
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	sdkstakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	sdkstakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app"
	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/testutil"
	"github.com/sedaprotocol/seda-chain/x/batching"
	batchingkeeper "github.com/sedaprotocol/seda-chain/x/batching/keeper"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
	dataproxykeeper "github.com/sedaprotocol/seda-chain/x/data-proxy/keeper"
	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"
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
	wasmstoragetestutil "github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper/testutil"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

const (
	bech32Prefix = "seda"
	bondDenom    = "aseda"
)

type fixture struct {
	*testutil.IntegationApp
	cdc               codec.Codec
	accountKeeper     authkeeper.AccountKeeper
	bankKeeper        bankkeeper.Keeper
	stakingKeeper     stakingkeeper.Keeper
	slashingKeeper    slashingkeeper.Keeper
	contractKeeper    wasmkeeper.PermissionedKeeper
	wasmKeeper        wasmkeeper.Keeper
	wasmStorageKeeper wasmstoragekeeper.Keeper
	tallyKeeper       tallykeeper.Keeper
	pubKeyKeeper      pubkeykeeper.Keeper
	batchingKeeper    batchingkeeper.Keeper
	mockViewKeeper    *wasmstoragetestutil.MockViewKeeper
	logBuf            *bytes.Buffer
}

func initFixture(tb testing.TB) *fixture {
	tb.Helper()

	tempDir := tb.TempDir()

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, sdkstakingtypes.StoreKey, wasmstoragetypes.StoreKey,
		wasmtypes.StoreKey, pubkeytypes.StoreKey, tallytypes.StoreKey, types.StoreKey, slashingtypes.StoreKey,
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
	stakingKeeper := stakingkeeper.NewKeeper(
		sdkStakingKeeper,
		addresscodec.NewBech32Codec(params.Bech32PrefixValAddr),
	)

	stakingParams := sdkstakingtypes.DefaultParams()
	stakingParams.BondDenom = bondDenom
	err := stakingKeeper.SetParams(ctx, stakingParams)
	require.NoError(tb, err)

	slashingKeeper := slashingkeeper.NewKeeper(
		cdc,
		nil,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		stakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	err = slashingKeeper.SetParams(ctx, slashingtypes.Params{
		SlashFractionDoubleSign: sdkmath.LegacyNewDec(1).Quo(sdkmath.LegacyNewDec(20)),
	})
	require.NoError(tb, err)

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
	require.NoError(tb, wasmKeeper.SetParams(ctx, wasmtypes.DefaultParams()))

	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(&wasmKeeper)

	ctrl := gomock.NewController(tb)
	viewKeeper := wasmstoragetestutil.NewMockViewKeeper(ctrl)

	wasmStorageKeeper := wasmstoragekeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[wasmstoragetypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authtypes.FeeCollectorName,
		nil,
		bankKeeper,
		stakingKeeper,
		contractKeeper,
	)

	pubKeyKeeper := pubkeykeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[pubkeytypes.StoreKey]),
		stakingKeeper,
		slashingKeeper,
		addresscodec.NewBech32Codec(params.Bech32PrefixValAddr),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
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
		runtime.NewKVStoreService(keys[types.StoreKey]),
		stakingKeeper,
		slashingKeeper,
		wasmStorageKeeper,
		pubKeyKeeper,
		contractKeeper,
		viewKeeper,
		addresscodec.NewBech32Codec(params.Bech32PrefixValAddr),
	)

	tallyKeeper := tallykeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[tallytypes.StoreKey]),
		wasmStorageKeeper,
		batchingKeeper,
		dataProxyKeeper,
		contractKeeper,
		viewKeeper,
		authority.String(),
	)

	authModule := auth.NewAppModule(cdc, accountKeeper, app.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, pubKeyKeeper)
	wasmStorageModule := wasmstorage.NewAppModule(cdc, *wasmStorageKeeper)
	tallyModule := tally.NewAppModule(cdc, tallyKeeper)
	pubKeyModule := pubkey.NewAppModule(cdc, pubKeyKeeper)
	batchingModule := batching.NewAppModule(cdc, batchingKeeper)

	integrationApp := testutil.NewIntegrationApp(ctx, logger, keys, cdc, router, map[string]appmodule.AppModule{
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
		slashingKeeper:    slashingKeeper,
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

func generateFirstBatch(t *testing.T, f *fixture, numValidators int) ([]sdk.ValAddress, []*ecdsa.PrivateKey, []sdkstakingtypes.Validator) {
	valAddrs, _, _ := addBatchSigningValidators(t, f, numValidators)

	// Create private keys for all validators and add them to the first batch
	privKeys := make([]*ecdsa.PrivateKey, numValidators)
	validatorAddrs := make([]sdk.ValAddress, numValidators)
	validators := make([]sdkstakingtypes.Validator, numValidators)
	validatorEntries := make([]types.ValidatorTreeEntry, numValidators)
	for i := 0; i < numValidators; i++ {
		privKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		privKeys[i] = privKey

		validatorAddr := sdk.ValAddress(valAddrs[i])
		validatorAddrs[i] = validatorAddr

		pubKeyBytes := crypto.FromECDSAPub(&privKey.PublicKey)
		ethAddr, err := utils.PubKeyToEthAddress(pubKeyBytes)
		require.NoError(t, err)
		validatorEntries[i] = types.ValidatorTreeEntry{
			ValidatorAddress: validatorAddr,
			EthAddress:       ethAddr,
		}

		validator, err := f.stakingKeeper.GetValidator(f.Context(), validatorAddr)
		require.NoError(t, err)
		validators[i] = validator

		consAddr, err := validator.GetConsAddr()
		require.NoError(t, err)

		f.slashingKeeper.SetValidatorSigningInfo(f.Context(), consAddr, slashingtypes.ValidatorSigningInfo{
			StartHeight: 1,
		})
	}

	// Generate a first batch and set it alongside the validators eth addresses
	batch := types.Batch{
		BatchId:     []byte("batch1"),
		BatchNumber: collections.DefaultSequenceStart,
		BlockHeight: 1,
	}
	err := f.batchingKeeper.SetNewBatch(f.Context(), batch, types.DataResultTreeEntries{}, validatorEntries)
	require.NoError(t, err)

	return validatorAddrs, privKeys, validators
}
