package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	sdkstakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper" // TO-DO rename
	sdkstakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"   // TO-DO rename

	"github.com/sedaprotocol/seda-chain/app"
	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/x/staking"
	stakingkeeper "github.com/sedaprotocol/seda-chain/x/staking/keeper"
	"github.com/sedaprotocol/seda-chain/x/vesting"
	"github.com/sedaprotocol/seda-chain/x/vesting/keeper"
	"github.com/sedaprotocol/seda-chain/x/vesting/types"
)

const Bech32Prefix = "seda"

var (
	bondDenom  = "aseda"
	coin100000 = sdk.NewInt64Coin(bondDenom, 100000)
	coin10000  = sdk.NewInt64Coin(bondDenom, 10000)
	coin7000   = sdk.NewInt64Coin(bondDenom, 7000)
	coin5000   = sdk.NewInt64Coin(bondDenom, 5000)
	coin2000   = sdk.NewInt64Coin(bondDenom, 2000)
	zeroCoins  sdk.Coins

	//
	funderAddr  = sdk.MustAccAddressFromBech32("seda1gujynygp0tkwzfpt0g7dv4829jwyk8f0yhp88d") // TO-DO cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5
	accountAddr = sdk.MustAccAddressFromBech32("seda1ucv5709wlf9jn84ynyjzyzeavwvurmdyxat26l") // TO-DO cosmos1x33fy6rusfprkntvjsfregss7rvsvyy4lkwrqu
	//
	testAddrs = []sdk.AccAddress{
		sdk.AccAddress([]byte("to0_________________")),
		sdk.AccAddress([]byte("to1_________________")),
		sdk.AccAddress([]byte("to2_________________")),
		sdk.AccAddress([]byte("to3_________________")),
		sdk.AccAddress([]byte("to4_________________")),
	}
)

type fixture struct {
	*app.IntegationApp
	cdc           codec.Codec
	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	stakingKeeper stakingkeeper.Keeper
}

func initFixture(tb testing.TB) *fixture {
	tb.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, sdkstakingtypes.StoreKey, types.StoreKey, //pooltypes.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}, vesting.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(tb)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, cmtproto.Header{Time: time.Now().UTC()}, true, logger)

	authority := authtypes.NewModuleAddress(types.ModuleName)

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

	// set default staking params
	stakingParams := sdkstakingtypes.DefaultParams()
	stakingParams.BondDenom = "aseda"
	err := stakingKeeper.SetParams(newCtx, stakingParams)
	require.NoError(tb, err)

	// Create MsgServiceRouter, but don't populate it before creating the gov
	// keeper.
	// router := baseapp.NewMsgServiceRouter()
	// router.SetInterfaceRegistry(cdc.InterfaceRegistry())
	authModule := auth.NewAppModule(cdc, accountKeeper, app.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, nil)
	vestingModule := vesting.NewAppModule(accountKeeper, bankKeeper, stakingKeeper)

	integrationApp := app.NewIntegrationApp(newCtx, logger, keys, cdc, map[string]appmodule.AppModule{
		authtypes.ModuleName:       authModule,
		banktypes.ModuleName:       bankModule,
		sdkstakingtypes.ModuleName: stakingModule,
		types.ModuleName:           vestingModule,
	})

	types.RegisterMsgServer(integrationApp.MsgServiceRouter(), keeper.NewMsgServerImpl(accountKeeper, bankKeeper, stakingKeeper))
	sdkstakingtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), sdkstakingkeeper.NewMsgServerImpl(sdkstakingKeeper))

	return &fixture{
		IntegationApp: integrationApp,
		cdc:           cdc,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		stakingKeeper: *stakingKeeper,
	}
}

func createValidators(t *testing.T, f *fixture, powers []int64) ([]sdk.AccAddress, []sdk.ValAddress) {
	t.Helper()
	addrs := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, f.Context(), 5, math.NewInt(30000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	pks := simtestutil.CreateTestPubKeys(5)

	val1, err := sdkstakingtypes.NewValidator(valAddrs[0].String(), pks[0], sdkstakingtypes.Description{})
	require.NoError(t, err)
	val2, err := sdkstakingtypes.NewValidator(valAddrs[1].String(), pks[1], sdkstakingtypes.Description{})
	require.NoError(t, err)
	val3, err := sdkstakingtypes.NewValidator(valAddrs[2].String(), pks[2], sdkstakingtypes.Description{})
	require.NoError(t, err)

	require.NoError(t, f.stakingKeeper.SetValidator(f.Context(), val1))
	require.NoError(t, f.stakingKeeper.SetValidator(f.Context(), val2))
	require.NoError(t, f.stakingKeeper.SetValidator(f.Context(), val3))
	require.NoError(t, f.stakingKeeper.SetValidatorByConsAddr(f.Context(), val1))
	require.NoError(t, f.stakingKeeper.SetValidatorByConsAddr(f.Context(), val2))
	require.NoError(t, f.stakingKeeper.SetValidatorByConsAddr(f.Context(), val3))
	require.NoError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.Context(), val1))
	require.NoError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.Context(), val2))
	require.NoError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.Context(), val3))

	_, _ = f.stakingKeeper.Delegate(f.Context(), addrs[0], f.stakingKeeper.TokensFromConsensusPower(f.Context(), powers[0]), sdkstakingtypes.Unbonded, val1, true)
	_, _ = f.stakingKeeper.Delegate(f.Context(), addrs[1], f.stakingKeeper.TokensFromConsensusPower(f.Context(), powers[1]), sdkstakingtypes.Unbonded, val2, true)
	_, _ = f.stakingKeeper.Delegate(f.Context(), addrs[2], f.stakingKeeper.TokensFromConsensusPower(f.Context(), powers[2]), sdkstakingtypes.Unbonded, val3, true)

	_, err = f.stakingKeeper.EndBlocker(f.Context())
	require.NoError(t, err)
	return addrs, valAddrs
}
