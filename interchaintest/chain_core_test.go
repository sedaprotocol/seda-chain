package interchaintest

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	genesisAmt = sdkmath.NewInt(10_000_000_000)

	testAddresses = []string{
		"seda1ucv5709wlf9jn84ynyjzyzeavwvurmdyxat26l",
		"seda10d07y265gmmuvt4z0w9aw880jnsr700jvvla4j", // gov module
		"seda1epj8499tu3cuxzrk3xrzzspl84pkafhx89sdp0",
		"seda1tvxrgdcpvj7ar9xhykzxwr7e8nztuec6pl8npz",
		"seda1s5jxphgva2dvx4dw7mzy6dae7rn8vjqhhl0t7v",
	}
)

func TestCoreSDKCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	t.Parallel()

	cosmos.SetSDKConfig(SedaCfg.Bech32Prefix)

	cfg := SedaCfg
	cfg.ModifyGenesis = cosmos.ModifyGenesis(getTestGenesis())

	chains := CreateChainsWithCustomConfig(t, numVals, numFullNodes, cfg)
	chain := chains[0].(*cosmos.CosmosChain)

	ic, ctx, _, _ := BuildAllChains(t, chains)

	t.Cleanup(func() {
		_ = ic.Close()
	})

	t.Run("authz", func(t *testing.T) {
		users := interchaintest.GetAndFundTestUsers(t, ctx, "default", genesisAmt, chain, chain, chain)
		testAuthz(ctx, t, chain, users)
	})

	t.Run("bank", func(t *testing.T) {
		users := interchaintest.GetAndFundTestUsers(t, ctx, "default", genesisAmt, chain, chain, chain)
		testBank(ctx, t, chain, users)
	})

	t.Run("distribution", func(t *testing.T) {
		users := interchaintest.GetAndFundTestUsers(t, ctx, "default", genesisAmt, chain, chain, chain)
		testDistribution(ctx, t, chain, users)
	})

	t.Run("feegrant", func(t *testing.T) {
		users := interchaintest.GetAndFundTestUsers(t, ctx, "default", genesisAmt, chain, chain, chain, chain)
		testFeeGrant(ctx, t, chain, users)
	})

	t.Run("auth", func(t *testing.T) {
		testAuth(ctx, t, chain)
	})

	t.Run("upgrade", func(t *testing.T) {
		testUpgrade(ctx, t, chain)
	})

	t.Run("staking", func(t *testing.T) {
		users := interchaintest.GetAndFundTestUsers(t, ctx, "default", genesisAmt, chain, chain, chain)
		testStaking(ctx, t, chain, users)
	})

	t.Run("slashing", func(t *testing.T) {
		testSlashing(ctx, t, chain)
	})
}

func testAuth(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain) {
	t.Helper()
	govAddr, err := chain.AuthQueryModuleAddress(ctx, "gov")
	require.NoError(t, err)
	require.NotEmpty(t, govAddr)

	govBz, err := chain.AccAddressFromBech32(govAddr)
	require.NoError(t, err)

	strAddr, err := chain.AuthAddressBytesToString(ctx, govBz)
	require.NoError(t, err)
	require.EqualValues(t, govAddr, strAddr)

	bz, err := chain.AuthAddressStringToBytes(ctx, strAddr)
	require.NoError(t, err)
	require.EqualValues(t, govBz, bz)

	p, err := chain.AuthQueryParams(ctx)
	require.NoError(t, err)
	require.NotNil(t, p)
	require.True(t, p.MaxMemoCharacters > 0)

	accs, err := chain.AuthQueryModuleAccounts(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, accs)

	bech32, err := chain.AuthQueryBech32Prefix(ctx)
	require.NoError(t, err)
	require.EqualValues(t, SedaCfg.Bech32Prefix, bech32)

	accInfo, err := chain.AuthQueryAccountInfo(ctx, govAddr)
	require.NoError(t, err)
	require.EqualValues(t, govAddr, accInfo.Address)
}

func testUpgrade(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain) {
	t.Helper()
	v, err := chain.UpgradeQueryAllModuleVersions(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, v)

	authority, err := chain.UpgradeQueryAuthority(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, authority)

	plan, err := chain.UpgradeQueryPlan(ctx)
	require.NoError(t, err)
	require.Nil(t, plan)

	_, err = chain.UpgradeQueryAppliedPlan(ctx, "")
	require.NoError(t, err)
}

func testAuthz(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain, users []ibc.Wallet) {
	t.Helper()
	granter := users[0].FormattedAddress()
	grantee := users[1].FormattedAddress()

	node := chain.GetNode()

	txRes, _ := node.AuthzGrant(ctx, users[0], grantee, "generic", "--msg-type", "/cosmos.bank.v1beta1.MsgSend")
	require.EqualValues(t, 0, txRes.Code)

	grants, err := chain.AuthzQueryGrants(ctx, granter, grantee, "")
	require.NoError(t, err)
	require.Len(t, grants, 1)
	require.EqualValues(t, grants[0].Authorization.TypeUrl, "/cosmos.authz.v1beta1.GenericAuthorization")
	require.Contains(t, string(grants[0].Authorization.Value), "/cosmos.bank.v1beta1.MsgSend")

	byGrantee, err := chain.AuthzQueryGrantsByGrantee(ctx, grantee, "")
	require.NoError(t, err)
	require.Len(t, byGrantee, 1)
	require.EqualValues(t, byGrantee[0].Granter, granter)
	require.EqualValues(t, byGrantee[0].Grantee, grantee)

	byGranter, err := chain.AuthzQueryGrantsByGranter(ctx, granter, "")
	require.NoError(t, err)
	require.Len(t, byGranter, 1)
	require.EqualValues(t, byGranter[0].Granter, granter)
	require.EqualValues(t, byGranter[0].Grantee, grantee)

	fmt.Printf("grants: %+v %+v %+v\n", grants, byGrantee, byGranter)

	balanceBefore, err := chain.GetBalance(ctx, granter, chain.Config().Denom)
	require.NoError(t, err)
	fmt.Printf("balanceBefore: %+v\n", balanceBefore)

	sendAmt := 1234

	nestedCmd := []string{
		chain.Config().Bin,
		"tx", "bank", "send", granter, grantee, fmt.Sprintf("%d%s", sendAmt, chain.Config().Denom),
		"--from", granter, "--generate-only",
		"--chain-id", chain.GetNode().Chain.Config().ChainID,
		"--node", chain.GetNode().Chain.GetRPCAddress(),
		"--home", chain.GetNode().HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"--output", "json",
		"--yes",
	}

	resp, err := node.AuthzExec(ctx, users[1], nestedCmd)
	require.NoError(t, err)
	require.EqualValues(t, 0, resp.Code)

	balanceAfter, err := chain.GetBalance(ctx, granter, chain.Config().Denom)
	require.NoError(t, err)

	fmt.Printf("balanceAfter: %+v\n", balanceAfter)
	require.EqualValues(t, balanceBefore.SubRaw(int64(sendAmt)), balanceAfter)
}

func testBank(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain, users []ibc.Wallet) {
	t.Helper()
	user0 := users[0].FormattedAddress()
	user1 := users[1].FormattedAddress()
	user2 := users[2].FormattedAddress()

	b2, err := chain.BankQueryBalance(ctx, user1, chain.Config().Denom)
	require.NoError(t, err)

	sendAmt := int64(1)
	_, err = sendTokens(ctx, chain, users[0], users[1], "", sendAmt)
	require.NoError(t, err)

	err = chain.GetNode().BankMultiSend(ctx, users[0].KeyName(), []string{user1, user2}, sdkmath.NewInt(sendAmt), chain.Config().Denom)
	require.NoError(t, err)

	b2New, err := chain.GetBalance(ctx, user1, chain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, b2.Add(sdkmath.NewInt(sendAmt*2)), b2New)

	b2All, err := chain.BankQueryAllBalances(ctx, user1)
	require.NoError(t, err)
	require.Equal(t, b2New, b2All.AmountOf(chain.Config().Denom))

	spendableBal, err := chain.BankQuerySpendableBalance(ctx, user0, chain.Config().Denom)
	require.NoError(t, err)

	spendableBals, err := chain.BankQuerySpendableBalances(ctx, user0)
	require.NoError(t, err)
	require.Equal(t, spendableBal.Amount, spendableBals.AmountOf(chain.Config().Denom))

	supply, err := chain.BankQueryTotalSupply(ctx)
	require.NoError(t, err)

	supplyOf, err := chain.BankQueryTotalSupplyOf(ctx, chain.Config().Denom)
	require.NoError(t, err)
	require.True(t, supplyOf.IsGTE(sdk.NewCoin(chain.Config().Denom, supply.AmountOf(chain.Config().Denom))))

	denomOwner, err := chain.BankQueryDenomOwners(ctx, chain.Config().Denom)
	require.NoError(t, err)

	found := false
	for _, owner := range denomOwner {
		if owner.Address == user0 {
			found = true
			break
		}
	}
	require.True(t, found)
}

func testDistribution(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain, users []ibc.Wallet) {
	t.Helper()
	var err error
	node := chain.GetNode()
	acc := authtypes.NewModuleAddress("distribution")
	require := require.New(t)

	vals, err := chain.StakingQueryValidators(ctx, stakingtypes.Bonded.String())
	require.NoError(err)
	fmt.Printf("validators: %+v\n", vals)

	del, err := chain.StakingQueryDelegationsTo(ctx, vals[0].OperatorAddress)
	require.NoError(err)

	delAddr := del[0].Delegation.DelegatorAddress
	valAddr := del[0].Delegation.ValidatorAddress

	newWithdrawAddr := testAddresses[0]

	t.Run("misc queries", func(t *testing.T) {
		slashes, err := chain.DistributionQueryValidatorSlashes(ctx, valAddr)
		require.NoError(err)
		require.EqualValues(0, len(slashes))

		valDistInfo, err := chain.DistributionQueryValidatorDistributionInfo(ctx, valAddr)
		require.NoError(err)
		fmt.Printf("valDistInfo: %+v\n", valDistInfo)
		require.EqualValues(1, valDistInfo.Commission.Len())

		valOutRewards, err := chain.DistributionQueryValidatorOutstandingRewards(ctx, valAddr)
		require.NoError(err)
		require.EqualValues(1, valOutRewards.Rewards.Len())

		params, err := chain.DistributionQueryParams(ctx)
		require.NoError(err)
		require.True(params.WithdrawAddrEnabled)

		comm, err := chain.DistributionQueryCommission(ctx, valAddr)
		require.NoError(err)
		require.EqualValues(chain.Config().Denom, comm.Commission[0].Denom)
	})

	t.Run("withdraw-all-rewards", func(t *testing.T) {
		err = node.StakingDelegate(ctx, users[2].KeyName(), valAddr, fmt.Sprintf("%d%s", uint64(100*math.Pow10(6)), chain.Config().Denom))
		require.NoError(err)

		before, err := chain.BankQueryBalance(ctx, acc.String(), chain.Config().Denom)
		require.NoError(err)
		fmt.Printf("before: %+v\n", before)

		err = node.DistributionWithdrawAllRewards(ctx, users[2].KeyName())
		require.NoError(err)

		after, err := chain.BankQueryBalance(ctx, acc.String(), chain.Config().Denom)
		require.NoError(err)
		fmt.Printf("after: %+v\n", after)
		require.True(after.GT(before))
	})

	t.Run("fund-pools", func(t *testing.T) {
		bal, err := chain.BankQueryBalance(ctx, acc.String(), chain.Config().Denom)
		require.NoError(err)
		fmt.Printf("CP balance: %+v\n", bal)

		amount := uint64(9_000 * math.Pow10(6))

		err = node.DistributionFundCommunityPool(ctx, users[0].KeyName(), fmt.Sprintf("%d%s", amount, chain.Config().Denom))
		require.NoError(err)

		err = node.DistributionFundValidatorRewardsPool(ctx, users[0].KeyName(), valAddr, fmt.Sprintf("%d%s", uint64(100*math.Pow10(6)), chain.Config().Denom))
		require.NoError(err)

		bal2, err := chain.BankQueryBalance(ctx, acc.String(), chain.Config().Denom)
		require.NoError(err)
		fmt.Printf("New CP balance: %+v\n", bal2) // 9147579661

		require.True(bal2.Sub(bal).GT(sdkmath.NewInt(int64(amount))))

		// queries
		coins, err := chain.DistributionQueryCommunityPool(ctx)
		require.NoError(err)
		require.True(coins.AmountOf(chain.Config().Denom).GT(sdkmath.LegacyNewDec(int64(amount))))
	})

	t.Run("set-custiom-withdraw-address", func(t *testing.T) {
		err = node.DistributionSetWithdrawAddr(ctx, users[0].KeyName(), newWithdrawAddr)
		require.NoError(err)

		withdrawAddr, err := chain.DistributionQueryDelegatorWithdrawAddress(ctx, users[0].FormattedAddress())
		require.NoError(err)
		require.EqualValues(withdrawAddr, newWithdrawAddr)
	})

	t.Run("delegator", func(t *testing.T) {
		delRewards, err := chain.DistributionQueryDelegationTotalRewards(ctx, delAddr)
		require.NoError(err)
		r := delRewards.Rewards[0]
		require.EqualValues(valAddr, r.ValidatorAddress)
		require.EqualValues(chain.Config().Denom, r.Reward[0].Denom)

		delegatorVals, err := chain.DistributionQueryDelegatorValidators(ctx, delAddr)
		require.NoError(err)
		require.EqualValues(valAddr, delegatorVals.Validators[0])

		rewards, err := chain.DistributionQueryRewards(ctx, delAddr, valAddr)
		require.NoError(err)
		require.EqualValues(1, rewards.Len())
	})
}

func testFeeGrant(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain, users []ibc.Wallet) {
	t.Helper()
	var err error
	node := chain.GetNode()

	denom := chain.Config().Denom

	t.Run("successful grant and queries", func(t *testing.T) {
		granter := users[0]
		grantee := users[1]

		err = node.FeeGrant(ctx, granter.KeyName(), grantee.FormattedAddress(), fmt.Sprintf("%d%s", 1000, chain.Config().Denom), []string{"/cosmos.bank.v1beta1.MsgSend"}, time.Now().Add(time.Hour*24*365))
		require.NoError(t, err)

		g, err := chain.FeeGrantQueryAllowance(ctx, granter.FormattedAddress(), grantee.FormattedAddress())
		require.NoError(t, err)
		fmt.Printf("g: %+v\n", g)
		require.EqualValues(t, granter.FormattedAddress(), g.Granter)
		require.EqualValues(t, grantee.FormattedAddress(), g.Grantee)
		require.EqualValues(t, "/cosmos.feegrant.v1beta1.AllowedMsgAllowance", g.Allowance.TypeUrl)
		require.Contains(t, string(g.Allowance.Value), "/cosmos.bank.v1beta1.MsgSend")

		all, err := chain.FeeGrantQueryAllowances(ctx, grantee.FormattedAddress())
		require.NoError(t, err)
		require.Len(t, all, 1)
		require.EqualValues(t, granter.FormattedAddress(), all[0].Granter)

		all2, err := chain.FeeGrantQueryAllowancesByGranter(ctx, granter.FormattedAddress())
		require.NoError(t, err)
		require.Len(t, all2, 1)
		require.EqualValues(t, grantee.FormattedAddress(), all2[0].Grantee)
	})

	t.Run("successful execution", func(t *testing.T) {
		granter2 := users[2]
		grantee2 := users[3]

		err = node.FeeGrant(ctx, granter2.KeyName(), grantee2.FormattedAddress(), fmt.Sprintf("%d%s", 100_000, denom), nil, time.Unix(0, 0))
		require.NoError(t, err)

		bal, err := chain.BankQueryBalance(ctx, granter2.FormattedAddress(), denom)
		require.NoError(t, err)

		fee := 500
		sendAmt := 501
		sendCoin := fmt.Sprintf("%d%s", sendAmt, denom)
		feeCoin := fmt.Sprintf("%d%s", fee, denom)

		_, err = node.ExecTx(ctx,
			grantee2.KeyName(), "bank", "send", grantee2.KeyName(), granter2.FormattedAddress(), sendCoin,
			"--fees", feeCoin, "--fee-granter", granter2.FormattedAddress(),
		)
		require.NoError(t, err)

		newBal, err := chain.BankQueryBalance(ctx, granter2.FormattedAddress(), denom)
		require.NoError(t, err)
		require.EqualValues(t, bal.AddRaw(int64(sendAmt-fee)), newBal)
	})
}

func testSlashing(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain) {
	t.Helper()
	p, err := chain.SlashingQueryParams(ctx)
	require.NoError(t, err)
	require.NotNil(t, p)

	infos, err := chain.SlashingQuerySigningInfos(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, infos)

	si, err := chain.SlashingQuerySigningInfo(ctx, infos[0].Address)
	require.NoError(t, err)
	require.NotNil(t, si)
}

func testStaking(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain, users []ibc.Wallet) {
	t.Helper()
	vals, err := chain.StakingQueryValidators(ctx, stakingtypes.Bonded.String())
	require.NoError(t, err)
	require.NotEmpty(t, vals)

	val := vals[0].OperatorAddress
	user := users[0].FormattedAddress()

	t.Run("query validators", func(t *testing.T) {
		valInfo, err := chain.StakingQueryValidator(ctx, val)
		require.NoError(t, err)
		require.EqualValues(t, val, valInfo.OperatorAddress)
		require.EqualValues(t, stakingtypes.Bonded.String(), valInfo.Status.String())

		del, err := chain.StakingQueryDelegationsTo(ctx, val)
		require.NoError(t, err)
		require.NotEmpty(t, del)

		del0 := del[0].Delegation.DelegatorAddress

		allDels, err := chain.StakingQueryDelegations(ctx, del0)
		require.NoError(t, err)
		require.NotEmpty(t, allDels)

		singleDel, err := chain.StakingQueryDelegation(ctx, val, del0)
		require.NoError(t, err)
		require.EqualValues(t, del0, singleDel.Delegation.DelegatorAddress)

		delVal, err := chain.StakingQueryDelegatorValidator(ctx, del0, val)
		require.NoError(t, err)
		require.True(t, delVal.OperatorAddress == val)

		delVals, err := chain.StakingQueryDelegatorValidators(ctx, del0)
		require.NoError(t, err)
		require.NotEmpty(t, delVals)
		require.True(t, delVals[0].OperatorAddress == val)
	})

	t.Run("misc", func(t *testing.T) {
		params, err := chain.StakingQueryParams(ctx)
		require.NoError(t, err)
		require.EqualValues(t, SedaDenom, params.BondDenom)

		pool, err := chain.StakingQueryPool(ctx)
		require.NoError(t, err)
		require.True(t, pool.BondedTokens.GT(sdkmath.NewInt(0)))

		height, err := chain.Height(ctx)
		require.NoError(t, err)

		searchHeight := int64(height - 1)

		hi, err := chain.StakingQueryHistoricalInfo(ctx, searchHeight)
		require.NoError(t, err)
		require.EqualValues(t, searchHeight, hi.Header.Height)
	})

	t.Run("delegations", func(t *testing.T) {
		node := chain.GetNode()

		err := node.StakingDelegate(ctx, users[0].KeyName(), val, "1000"+chain.Config().Denom)
		require.NoError(t, err)

		dels, err := chain.StakingQueryDelegations(ctx, users[0].FormattedAddress())
		require.NoError(t, err)
		found := false
		for _, d := range dels {
			if d.Balance.Amount.Equal(sdkmath.NewInt(1000)) {
				found = true
				break
			}
		}
		require.True(t, found)

		err = node.StakingUnbond(ctx, users[0].KeyName(), val, "25"+chain.Config().Denom)
		require.NoError(t, err)

		unbonding, err := chain.StakingQueryUnbondingDelegation(ctx, user, val)
		require.NoError(t, err)
		require.EqualValues(t, user, unbonding.DelegatorAddress)
		require.EqualValues(t, val, unbonding.ValidatorAddress)

		height := unbonding.Entries[0].CreationHeight

		unbondings, err := chain.StakingQueryUnbondingDelegations(ctx, user)
		require.NoError(t, err)
		require.NotEmpty(t, unbondings)
		require.EqualValues(t, user, unbondings[0].DelegatorAddress)

		unbondingsFrom, err := chain.StakingQueryUnbondingDelegationsFrom(ctx, val)
		require.NoError(t, err)
		require.NotEmpty(t, unbondingsFrom)
		require.EqualValues(t, user, unbondingsFrom[0].DelegatorAddress)

		err = node.StakingCancelUnbond(ctx, user, val, "25"+chain.Config().Denom, height)
		require.NoError(t, err)

		unbondings, err = chain.StakingQueryUnbondingDelegations(ctx, user)
		require.NoError(t, err)
		require.Empty(t, unbondings)
	})
}

func sendTokens(ctx context.Context, chain *cosmos.CosmosChain, from, to ibc.Wallet, token string, amount int64) (ibc.WalletAmount, error) {
	if token == "" {
		token = chain.Config().Denom
	}

	sendAmt := ibc.WalletAmount{
		Address: to.FormattedAddress(),
		Denom:   token,
		Amount:  sdkmath.NewInt(amount),
	}
	err := chain.SendFunds(ctx, from.KeyName(), sendAmt)
	return sendAmt, err
}
