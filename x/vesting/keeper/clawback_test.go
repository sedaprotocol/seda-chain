package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	sdkstakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/x/vesting/types"
)

func TestClawback(t *testing.T) {
	f := initFixture(t)
	f.bankKeeper.SetSendEnabled(f.Context(), "aseda", true)
	err := banktestutil.FundAccount(f.Context(), f.bankKeeper, funderAddr, sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 500000)))
	require.NoError(t, err)

	_, valAddrs, valPks := createValidators(t, f, []int64{5, 5, 5})
	require.NoError(t, err)

	testCases := []struct {
		testName                string
		funder                  sdk.AccAddress
		recipient               sdk.AccAddress
		vestingTime             int64
		timeUntilClawback       int64
		originalVesting         sdk.Coin
		delegations             []sdk.Coin // to val0 and val1
		undelegations           []sdk.Coin // from val0 and val1
		redelegations           []sdk.Coin // (val0 -> val1) and (val1 -> val0)
		expClawedUnbonded       sdk.Coins
		expClawedUnbonding      sdk.Coins
		expClawedBonded         sdk.Coins
		slashFractions          []math.LegacyDec // on val0 and val1
		recipientFinalSpendable sdk.Coins
	}{
		{
			testName:                "clawback from unbonded",
			funder:                  sdk.MustAccAddressFromBech32("seda1gujynygp0tkwzfpt0g7dv4829jwyk8f0yhp88d"),
			recipient:               testAddrs[0],
			vestingTime:             100,
			timeUntilClawback:       30,
			originalVesting:         sdk.NewInt64Coin(bondDenom, 10000),
			delegations:             []sdk.Coin{sdk.NewInt64Coin(bondDenom, 0), sdk.NewInt64Coin(bondDenom, 0)},
			undelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 0), sdk.NewInt64Coin(bondDenom, 0)},
			redelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 0), sdk.NewInt64Coin(bondDenom, 0)},
			expClawedUnbonded:       sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 7000)),
			expClawedUnbonding:      zeroCoins,
			expClawedBonded:         zeroCoins,
			slashFractions:          []math.LegacyDec{math.LegacyZeroDec(), math.LegacyZeroDec()},
			recipientFinalSpendable: sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 3000)),
		},
		{
			testName:                "clawback from unbonded and bonded",
			funder:                  sdk.MustAccAddressFromBech32("seda1gujynygp0tkwzfpt0g7dv4829jwyk8f0yhp88d"),
			recipient:               testAddrs[1],
			vestingTime:             100,
			timeUntilClawback:       30,
			originalVesting:         sdk.NewInt64Coin(bondDenom, 10000),
			delegations:             []sdk.Coin{sdk.NewInt64Coin(bondDenom, 5000), sdk.NewInt64Coin(bondDenom, 0)},
			undelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 0), sdk.NewInt64Coin(bondDenom, 0)},
			redelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 0), sdk.NewInt64Coin(bondDenom, 0)},
			expClawedUnbonded:       sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 5000)),
			expClawedUnbonding:      zeroCoins,
			expClawedBonded:         sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 2000)),
			slashFractions:          []math.LegacyDec{math.LegacyZeroDec(), math.LegacyZeroDec()},
			recipientFinalSpendable: sdk.NewCoins(),
		},
		{
			testName:                "clawback from bonded",
			funder:                  sdk.MustAccAddressFromBech32("seda1gujynygp0tkwzfpt0g7dv4829jwyk8f0yhp88d"),
			recipient:               testAddrs[2],
			vestingTime:             100,
			timeUntilClawback:       60,
			originalVesting:         sdk.NewInt64Coin(bondDenom, 27500),
			delegations:             []sdk.Coin{sdk.NewInt64Coin(bondDenom, 27500), sdk.NewInt64Coin(bondDenom, 0)},
			undelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 0), sdk.NewInt64Coin(bondDenom, 0)},
			redelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 0), sdk.NewInt64Coin(bondDenom, 0)},
			expClawedUnbonded:       zeroCoins,
			expClawedUnbonding:      zeroCoins,
			expClawedBonded:         sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 11000)),
			slashFractions:          []math.LegacyDec{math.LegacyZeroDec(), math.LegacyZeroDec()},
			recipientFinalSpendable: sdk.NewCoins(),
		},
		{
			testName:                "clawback from unbonding",
			funder:                  sdk.MustAccAddressFromBech32("seda1gujynygp0tkwzfpt0g7dv4829jwyk8f0yhp88d"),
			recipient:               testAddrs[3],
			vestingTime:             50000,
			timeUntilClawback:       30000,
			originalVesting:         sdk.NewInt64Coin(bondDenom, 27500),
			delegations:             []sdk.Coin{sdk.NewInt64Coin(bondDenom, 27500), sdk.NewInt64Coin(bondDenom, 0)},
			undelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 27500), sdk.NewInt64Coin(bondDenom, 0)},
			redelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 0), sdk.NewInt64Coin(bondDenom, 0)},
			expClawedUnbonded:       zeroCoins,
			expClawedUnbonding:      sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 11000)),
			expClawedBonded:         zeroCoins,
			slashFractions:          []math.LegacyDec{math.LegacyZeroDec(), math.LegacyZeroDec()},
			recipientFinalSpendable: sdk.NewCoins(),
		},
		{
			testName:                "clawback from unbonded, unbonding, and bonded",
			funder:                  sdk.MustAccAddressFromBech32("seda1gujynygp0tkwzfpt0g7dv4829jwyk8f0yhp88d"),
			recipient:               testAddrs[4],
			vestingTime:             750000,
			timeUntilClawback:       600000,
			originalVesting:         sdk.NewInt64Coin(bondDenom, 13000),
			delegations:             []sdk.Coin{sdk.NewInt64Coin(bondDenom, 10000), sdk.NewInt64Coin(bondDenom, 2000)},
			undelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 400), sdk.NewInt64Coin(bondDenom, 100)},
			redelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 0), sdk.NewInt64Coin(bondDenom, 0)},
			expClawedUnbonded:       sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 1000)),
			expClawedUnbonding:      sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 500)),
			expClawedBonded:         sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 1100)),
			slashFractions:          []math.LegacyDec{math.LegacyZeroDec(), math.LegacyZeroDec()},
			recipientFinalSpendable: sdk.NewCoins(),
		},
		{
			testName:                "clawback from unbonded and bonded with slashing",
			funder:                  sdk.MustAccAddressFromBech32("seda1gujynygp0tkwzfpt0g7dv4829jwyk8f0yhp88d"),
			recipient:               testAddrs[5],
			vestingTime:             100,
			timeUntilClawback:       30,
			originalVesting:         sdk.NewInt64Coin(bondDenom, 10000),
			delegations:             []sdk.Coin{sdk.NewInt64Coin(bondDenom, 5000), sdk.NewInt64Coin(bondDenom, 0)},
			undelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 0), sdk.NewInt64Coin(bondDenom, 0)},
			redelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 0), sdk.NewInt64Coin(bondDenom, 0)},
			expClawedUnbonded:       sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 5000)),
			expClawedUnbonding:      zeroCoins,
			expClawedBonded:         sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 2000)),
			slashFractions:          []math.LegacyDec{math.LegacyNewDecWithPrec(5, 2), math.LegacyZeroDec()}, // 0.05 and 0
			recipientFinalSpendable: sdk.NewCoins(),
		},
		{
			testName:                "clawback from unbonded, unbonding, and bonded with slashing",
			funder:                  sdk.MustAccAddressFromBech32("seda1gujynygp0tkwzfpt0g7dv4829jwyk8f0yhp88d"),
			recipient:               testAddrs[6],
			vestingTime:             750000,
			timeUntilClawback:       600000,
			originalVesting:         sdk.NewInt64Coin(bondDenom, 13000),
			delegations:             []sdk.Coin{sdk.NewInt64Coin(bondDenom, 10000), sdk.NewInt64Coin(bondDenom, 2000)},
			undelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 400), sdk.NewInt64Coin(bondDenom, 100)},
			redelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 0), sdk.NewInt64Coin(bondDenom, 0)},
			expClawedUnbonded:       sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 1000)),
			expClawedUnbonding:      sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 180)),
			expClawedBonded:         sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 1420)),
			slashFractions:          []math.LegacyDec{math.LegacyNewDecWithPrec(8, 1), math.LegacyZeroDec()}, // 0.8 and 0
			recipientFinalSpendable: sdk.NewCoins(),
		},
		{
			testName:                "clawback from redelegation with slashing",
			funder:                  sdk.MustAccAddressFromBech32("seda1gujynygp0tkwzfpt0g7dv4829jwyk8f0yhp88d"),
			recipient:               testAddrs[7],
			vestingTime:             50000,
			timeUntilClawback:       30000,
			originalVesting:         sdk.NewInt64Coin(bondDenom, 27500),
			delegations:             []sdk.Coin{sdk.NewInt64Coin(bondDenom, 27500), sdk.NewInt64Coin(bondDenom, 0)},
			undelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 0), sdk.NewInt64Coin(bondDenom, 0)},
			redelegations:           []sdk.Coin{sdk.NewInt64Coin(bondDenom, 27500), sdk.NewInt64Coin(bondDenom, 0)},
			expClawedUnbonded:       zeroCoins,
			expClawedUnbonding:      zeroCoins,
			expClawedBonded:         sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 9625)),
			slashFractions:          []math.LegacyDec{math.LegacyZeroDec(), math.LegacyNewDecWithPrec(65, 2)}, // 0 and 0.65
			recipientFinalSpendable: sdk.NewCoins(),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			f.AddBlock()

			// 1. create clawback continuous vesting account
			createVestingMsg := &types.MsgCreateVestingAccount{
				FromAddress: tc.funder.String(),
				ToAddress:   tc.recipient.String(),
				Amount:      sdk.NewCoins(tc.originalVesting),
				EndTime:     f.Context().BlockTime().Unix() + tc.vestingTime,
			}
			_, err = f.RunMsg(createVestingMsg)
			require.NoError(t, err)

			// 2. delegate
			if tc.delegations[0].IsPositive() {
				delegateMsg := &sdkstakingtypes.MsgDelegate{
					DelegatorAddress: tc.recipient.String(),
					ValidatorAddress: valAddrs[0].String(),
					Amount:           tc.delegations[0],
				}
				_, err = f.RunMsg(delegateMsg)
				require.NoError(t, err)
			}
			if tc.delegations[1].IsPositive() {
				delegateMsg := &sdkstakingtypes.MsgDelegate{
					DelegatorAddress: tc.recipient.String(),
					ValidatorAddress: valAddrs[1].String(),
					Amount:           tc.delegations[1],
				}
				_, err = f.RunMsg(delegateMsg)
				require.NoError(t, err)
			}

			// 3. initiate unbonding after some time
			if tc.undelegations[0].IsPositive() {
				undelegateMsg := &sdkstakingtypes.MsgUndelegate{
					DelegatorAddress: tc.recipient.String(),
					ValidatorAddress: valAddrs[0].String(),
					Amount:           tc.undelegations[0],
				}
				_, err = f.RunMsg(undelegateMsg)
				require.NoError(t, err)
			}
			if tc.undelegations[1].IsPositive() {
				undelegateMsg := &sdkstakingtypes.MsgUndelegate{
					DelegatorAddress: tc.recipient.String(),
					ValidatorAddress: valAddrs[1].String(),
					Amount:           tc.undelegations[1],
				}
				_, err = f.RunMsg(undelegateMsg)
				require.NoError(t, err)
			}

			if tc.redelegations[0].IsPositive() {
				redelegateMsg := &sdkstakingtypes.MsgBeginRedelegate{
					DelegatorAddress:    tc.recipient.String(),
					ValidatorSrcAddress: valAddrs[0].String(),
					ValidatorDstAddress: valAddrs[1].String(),
					Amount:              tc.redelegations[0],
				}
				_, err = f.RunMsg(redelegateMsg)
				require.NoError(t, err)
			}
			if tc.redelegations[1].IsPositive() {
				redelegateMsg := &sdkstakingtypes.MsgBeginRedelegate{
					DelegatorAddress:    tc.recipient.String(),
					ValidatorSrcAddress: valAddrs[1].String(),
					ValidatorDstAddress: valAddrs[0].String(),
					Amount:              tc.redelegations[1],
				}
				_, err = f.RunMsg(redelegateMsg)
				require.NoError(t, err)
			}

			// possible slashing
			if tc.slashFractions[0].IsPositive() {
				_, err = f.stakingKeeper.Slash(f.Context(), sdk.ConsAddress(valPks[0].Address()), f.Context().BlockHeight()-1, 5, tc.slashFractions[0])
				require.NoError(t, err)
			}
			if tc.slashFractions[1].IsPositive() {
				_, err = f.stakingKeeper.Slash(f.Context(), sdk.ConsAddress(valPks[1].Address()), f.Context().BlockHeight()-1, 5, tc.slashFractions[1])
				require.NoError(t, err)
			}

			_, err = f.stakingKeeper.EndBlocker(f.Context())
			require.NoError(t, err)
			f.AddBlock()

			// 4. clawback after some time
			f.AddTime(tc.timeUntilClawback)

			clawbackMsg := &types.MsgClawback{
				FunderAddress:  tc.funder.String(),
				AccountAddress: tc.recipient.String(),
			}
			res, err := f.RunMsg(clawbackMsg)
			require.NoError(t, err)

			result := types.MsgClawbackResponse{}
			err = f.cdc.Unmarshal(res.Value, &result)
			require.NoError(t, err)

			require.Equal(t, tc.expClawedUnbonded, result.ClawedUnbonded)
			require.Equal(t, tc.expClawedUnbonding, result.ClawedUnbonding)
			require.Equal(t, tc.expClawedBonded, result.ClawedBonded)

			//
			recipientSpendable := f.bankKeeper.SpendableCoins(f.Context(), tc.recipient)
			require.Equal(t, tc.recipientFinalSpendable, recipientSpendable)
		})
	}
}
