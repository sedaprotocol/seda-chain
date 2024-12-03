package keeper_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdkintegration "github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
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
	"github.com/sedaprotocol/seda-chain/integration"
	"github.com/sedaprotocol/seda-chain/x/pubkey"
	"github.com/sedaprotocol/seda-chain/x/pubkey/keeper"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
	"github.com/sedaprotocol/seda-chain/x/staking"
	stakingkeeper "github.com/sedaprotocol/seda-chain/x/staking/keeper"
	"github.com/sedaprotocol/seda-chain/x/vesting"
)

const (
	bech32Prefix = "seda"
	bondDenom    = "aseda"
)

var (
	zeroCoins  sdk.Coins
	funderAddr = sdk.MustAccAddressFromBech32("seda1gujynygp0tkwzfpt0g7dv4829jwyk8f0yhp88d")
	testAddrs  = []sdk.AccAddress{
		sdk.AccAddress([]byte("to0_________________")),
		sdk.AccAddress([]byte("to1_________________")),
		sdk.AccAddress([]byte("to2_________________")),
		sdk.AccAddress([]byte("to3_________________")),
		sdk.AccAddress([]byte("to4_________________")),
		sdk.AccAddress([]byte("to5_________________")),
		sdk.AccAddress([]byte("to6_________________")),
		sdk.AccAddress([]byte("to7_________________")),
		sdk.AccAddress([]byte("to8_________________")),
		sdk.AccAddress([]byte("to9_________________")),
	}
)

type fixture struct {
	*integration.IntegationApp
	cdc           codec.Codec
	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	stakingKeeper stakingkeeper.Keeper
	keeper        keeper.Keeper
}

func initFixture(tb testing.TB) *fixture {
	tb.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, sdkstakingtypes.StoreKey, types.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}, vesting.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(tb)
	cms := sdkintegration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, cmtproto.Header{Time: time.Now().UTC()}, true, logger)

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
	err := stakingKeeper.SetParams(newCtx, stakingParams)
	require.NoError(tb, err)

	slashingKeeper := slashingkeeper.NewKeeper(
		cdc,
		nil,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		stakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	pubkeyKeeper := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[types.StoreKey]),
		stakingKeeper,
		slashingKeeper,
		addresscodec.NewBech32Codec(params.Bech32PrefixValAddr),
	)

	authModule := auth.NewAppModule(cdc, accountKeeper, app.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, nil)
	pubkeyModule := pubkey.NewAppModule(cdc, *pubkeyKeeper)

	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc, map[string]appmodule.AppModule{
		authtypes.ModuleName:       authModule,
		banktypes.ModuleName:       bankModule,
		sdkstakingtypes.ModuleName: stakingModule,
		types.ModuleName:           pubkeyModule,
	})

	types.RegisterMsgServer(integrationApp.MsgServiceRouter(), keeper.NewMsgServerImpl(*pubkeyKeeper))
	sdkstakingtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), sdkstakingkeeper.NewMsgServerImpl(sdkstakingKeeper))

	return &fixture{
		IntegationApp: integrationApp,
		cdc:           cdc,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		stakingKeeper: *stakingKeeper,
		keeper:        *pubkeyKeeper,
	}
}

func createValidators(t *testing.T, f *fixture, powers []int64) ([]sdk.AccAddress, []sdk.ValAddress, []cryptotypes.PubKey) {
	t.Helper()
	addrs := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, f.Context(), len(powers), math.NewIntFromUint64(1e19))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	valPks := simtestutil.CreateTestPubKeys(len(powers))

	for i := 0; i < len(powers); i++ {
		val, err := sdkstakingtypes.NewValidator(valAddrs[i].String(), valPks[i], sdkstakingtypes.Description{})
		require.NoError(t, err)

		require.NoError(t, f.stakingKeeper.SetValidator(f.Context(), val))
		require.NoError(t, f.stakingKeeper.SetValidatorByConsAddr(f.Context(), val))
		require.NoError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.Context(), val))

		_, err = f.stakingKeeper.Delegate(f.Context(), addrs[i], f.stakingKeeper.TokensFromConsensusPower(f.Context(), powers[i]), sdkstakingtypes.Unbonded, val, true)
		require.NoError(t, err)
	}

	_, err := f.stakingKeeper.EndBlocker(f.Context())
	require.NoError(t, err)
	return addrs, valAddrs, valPks
}

func TestEndBlock(t *testing.T) {
	f := initFixture(t)
	ctx := f.Context()

	f.keeper.InitGenesis(ctx, *types.DefaultGenesisState())

	_, valAddrs, _ := createValidators(t, f, []int64{1, 3, 5, 7, 2, 1}) // 1+3+5+7+2+1 = 19
	pubKeys := generatePubKeys(t, 6)

	// Check for start of activation process.
	var expectedActivationHeight int64 = types.DefaultActivationHeight
	for i := range valAddrs[:len(valAddrs)-1] {
		err := f.keeper.SetValidatorKeyAtIndex(ctx, valAddrs[i], utils.SEDAKeyIndexSecp256k1, pubKeys[i])
		require.NoError(t, err)
		err = f.keeper.EndBlock(ctx)
		require.NoError(t, err)

		if i >= 3 {
			expectedActivationHeight = ctx.BlockHeight() + keeper.ActivationLag
		}
		scheme, err := f.keeper.GetProvingScheme(ctx, utils.SEDAKeyIndexSecp256k1)
		require.NoError(t, err)
		require.Equal(t, expectedActivationHeight, scheme.ActivationHeight)
		require.Equal(t, scheme.IsActivated, false)
	}

	// Check for successful activation.
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + keeper.ActivationLag)
	err := f.keeper.EndBlock(ctx)
	require.NoError(t, err)
	scheme, err := f.keeper.GetProvingScheme(ctx, utils.SEDAKeyIndexSecp256k1)
	require.NoError(t, err)
	require.Equal(t, scheme.IsActivated, true)

	for i := range valAddrs {
		val, err := f.stakingKeeper.GetValidator(ctx, valAddrs[i])
		require.NoError(t, err)
		if i == len(valAddrs)-1 {
			require.Equal(t, val.Jailed, true)
		} else {
			require.Equal(t, val.Jailed, false)
		}
	}
}

func generatePubKeys(t *testing.T, num int) [][]byte {
	t.Helper()
	var pubKeys [][]byte
	for i := 0; i < num; i++ {
		privKey, err := ecdsa.GenerateKey(ethcrypto.S256(), rand.Reader)
		if err != nil {
			panic(fmt.Sprintf("failed to generate secp256k1 private key: %v", err))
		}
		pubKeys = append(pubKeys, elliptic.Marshal(privKey.PublicKey, privKey.PublicKey.X, privKey.PublicKey.Y))
	}
	return pubKeys
}
