package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	sdkstakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

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
	fromAddr  = sdk.AccAddress([]byte("from1________________"))
	to1Addr   = sdk.AccAddress([]byte("to1__________________"))
	to2Addr   = sdk.AccAddress([]byte("to2__________________"))
	to3Addr   = sdk.AccAddress([]byte("to3__________________"))
	bondDenom = "aseda"
	fooCoin   = sdk.NewInt64Coin(bondDenom, 10000)
	barCoin   = sdk.NewInt64Coin(bondDenom, 7000)
	zeroCoin  = sdk.NewInt64Coin(bondDenom, 0)
)

type fixture struct {
	// ctx sdk.Context

	app *app.IntegationApp
	// queryClient       v1.QueryClient
	// legacyQueryClient v1beta1.QueryClient
	cdc codec.Codec

	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	stakingKeeper stakingkeeper.Keeper
}

func initFixture(tb testing.TB) *fixture {
	tb.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey, types.StoreKey, //pooltypes.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}, vesting.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(tb)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, cmtproto.Header{Time: time.Now().UTC()}, true, logger)

	authority := authtypes.NewModuleAddress(types.ModuleName)

	maccPerms := map[string][]string{
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		types.ModuleName:               {authtypes.Burner},
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

	sdkstakingKeeper := sdkstakingkeeper.NewKeeper(cdc, runtime.NewKVStoreService(keys[stakingtypes.StoreKey]), accountKeeper, bankKeeper, authority.String(), addresscodec.NewBech32Codec(params.Bech32PrefixValAddr), addresscodec.NewBech32Codec(params.Bech32PrefixConsAddr))
	stakingKeeper := stakingkeeper.NewKeeper(sdkstakingKeeper)

	// set default staking params
	stakingParams := stakingtypes.DefaultParams()
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
		authtypes.ModuleName:    authModule,
		banktypes.ModuleName:    bankModule,
		stakingtypes.ModuleName: stakingModule,
		types.ModuleName:        vestingModule,
	})

	// sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	// msgSrvr := keeper.NewMsgServerImpl(govKeeper)
	// legacyMsgSrvr := keeper.NewLegacyMsgServerImpl(authority.String(), msgSrvr)

	// // Register MsgServer and QueryServer
	types.RegisterMsgServer(integrationApp.MsgServiceRouter(), keeper.NewMsgServerImpl(accountKeeper, bankKeeper, stakingKeeper))
	// v1beta1.RegisterMsgServer(router, legacyMsgSrvr)

	// v1.RegisterQueryServer(integrationApp.QueryHelper(), keeper.NewQueryServer(govKeeper))
	// v1beta1.RegisterQueryServer(integrationApp.QueryHelper(), keeper.NewLegacyQueryServer(govKeeper))

	// queryClient := v1.NewQueryClient(integrationApp.QueryHelper())
	// legacyQueryClient := v1beta1.NewQueryClient(integrationApp.QueryHelper())

	return &fixture{
		app: integrationApp,
		cdc: cdc,
		// ctx: sdkCtx,
		// queryClient:       queryClient,
		// legacyQueryClient: legacyQueryClient,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		stakingKeeper: *stakingKeeper,
		// govKeeper:         govKeeper,
	}
}

func TestClawbackContinuousVesting(t *testing.T) {
	f := initFixture(t)

	// _ = f.ctx

	funderAddr := sdk.MustAccAddressFromBech32("seda1gujynygp0tkwzfpt0g7dv4829jwyk8f0yhp88d")  // TO-DO cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5
	accountAddr := sdk.MustAccAddressFromBech32("seda1ucv5709wlf9jn84ynyjzyzeavwvurmdyxat26l") // TO-DO cosmos1x33fy6rusfprkntvjsfregss7rvsvyy4lkwrqu

	originalVesting := sdk.NewCoins(fooCoin)
	endTime := f.app.Context().BlockTime().Unix() + 100
	// bondedAmt := math.NewInt(0)
	// unbondingAmt := math.NewInt(0)
	// unbonded := originalVesting // TO-DO why? GetAllBalances

	// msg := &types.MsgClawback{
	// 	FunderAddress:  funderAddr.String(),
	// 	AccountAddress: accountAddr.String(),
	// }

	// funderAcc := authtypes.NewBaseAccountWithAddress(funderAddr)
	// baseAccount := authtypes.NewBaseAccountWithAddress(accountAddr)
	// baseVestingAccount, err := sdkvestingtypes.NewBaseVestingAccount(baseAccount, originalVesting, endTime)
	// require.NoError(t, err)
	// vestingAccount := types.NewClawbackContinuousVestingAccountRaw(baseVestingAccount, f.ctx.BlockTime().Unix(), msg.FunderAddress)
	// f.accountKeeper.SetAccount(f.ctx, vestingAccount)

	f.bankKeeper.SetSendEnabled(f.app.Context(), "aseda", true)

	err := banktestutil.FundAccount(f.app.Context(), f.bankKeeper, funderAddr, sdk.NewCoins(fooCoin))
	require.NoError(t, err)

	// 1. create clawback continuous vesting account
	msg := &types.MsgCreateVestingAccount{
		FromAddress: funderAddr.String(),
		ToAddress:   accountAddr.String(),
		Amount:      originalVesting,
		EndTime:     endTime,
	}
	res, err := f.app.RunMsg(msg)
	require.NoError(t, err)
	fmt.Println(res)

	// 2. clawback after some time
	f.app.AddTime(30)
	clawbackMsg := &types.MsgClawback{
		FunderAddress:  funderAddr.String(),
		AccountAddress: accountAddr.String(), // TO-DO rename to RecipientAddress?
	}
	res, err = f.app.RunMsg(clawbackMsg)
	require.NoError(t, err)

	result := types.MsgClawbackResponse{}
	err = f.cdc.Unmarshal(res.Value, &result)
	require.NoError(t, err)
	fmt.Println(result)

	require.Equal(t, sdk.NewCoins(barCoin), result.Coins)
	// createValidators(t, f, []int64{5, 5, 5})
}

// Test when 0 time has passed
// toXfer seems to be 0
