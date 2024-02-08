package keeper

import (
	"context"
	"fmt"
	"math"

	"github.com/hashicorp/go-metrics"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/sedaprotocol/seda-chain/x/vesting/types"
)

type msgServer struct {
	ak types.AccountKeeper
	bk types.BankKeeper
	sk types.StakingKeeper
}

// NewMsgServerImpl returns an implementation of the vesting MsgServer interface,
// wrapping the corresponding AccountKeeper and BankKeeper.
func NewMsgServerImpl(ak types.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper) types.MsgServer {
	return &msgServer{ak: ak, bk: bk, sk: sk}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) CreateVestingAccount(goCtx context.Context, msg *types.MsgCreateVestingAccount) (*types.MsgCreateVestingAccountResponse, error) {
	from, err := m.ak.AddressCodec().StringToBytes(msg.FromAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'from' address: %s", err)
	}

	to, err := m.ak.AddressCodec().StringToBytes(msg.ToAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'to' address: %s", err)
	}

	if err := validateAmount(msg.Amount); err != nil {
		return nil, err
	}

	if msg.EndTime <= 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid end time")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := m.bk.IsSendEnabledCoins(ctx, msg.Amount...); err != nil {
		return nil, err
	}

	if m.bk.BlockedAddr(to) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", msg.ToAddress)
	}

	if acc := m.ak.GetAccount(ctx, to); acc != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "account %s already exists", msg.ToAddress)
	}

	baseAccount := authtypes.NewBaseAccountWithAddress(to)
	baseAccount = m.ak.NewAccount(ctx, baseAccount).(*authtypes.BaseAccount)
	baseVestingAccount, err := sdkvestingtypes.NewBaseVestingAccount(baseAccount, msg.Amount.Sort(), msg.EndTime)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	vestingAccount := types.NewClawbackContinuousVestingAccountRaw(baseVestingAccount, ctx.BlockTime().Unix(), msg.FromAddress)

	m.ak.SetAccount(ctx, vestingAccount)

	defer func() {
		telemetry.IncrCounter(1, "new", "account")

		for _, a := range msg.Amount {
			if a.Amount.IsInt64() {
				telemetry.SetGaugeWithLabels(
					[]string{"tx", "msg", "create_vesting_account"},
					float32(a.Amount.Int64()),
					[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
				)
			}
		}
	}()

	if err = m.bk.SendCoins(ctx, from, to, msg.Amount); err != nil {
		return nil, err
	}

	return &types.MsgCreateVestingAccountResponse{}, nil
}

// Clawback removes the unvested amount from a ClawbackVestingAccount.
//
// Checks performed on the ValidateBasic include:
//   - funder and vesting addresses are correct bech32 format
//   - if destination address is not empty it is also correct bech32 format
func (m msgServer) Clawback(goCtx context.Context, msg *types.MsgClawback) (*types.MsgClawbackResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// NOTE: errors checked during msg validation
	vestingAccAddr := sdk.MustAccAddressFromBech32(msg.AccountAddress)
	funderAddr := sdk.MustAccAddressFromBech32(msg.FunderAddress)

	// // NOTE: we check the destination address only for the case where it's not sent from the
	// // authority account, because in that case the destination address is hardcored to the
	// // community pool address anyway (see further below).
	// if bk.BlockedAddr(dest) {
	// 	return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized,
	// 		"%s is a blocked address and not allowed to receive funds", msg.DestAddress,
	// 	)
	// }

	// retrieve the vesting account and perform preliminary checks
	acc := m.ak.GetAccount(ctx, vestingAccAddr)
	if acc == nil {
		// TO-DO what to return?
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "account at address '%s' does not exist", vestingAccAddr.String())
	}
	vestingAccount, isClawback := acc.(*types.ClawbackContinuousVestingAccount)
	if !isClawback {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "account %s is not of type ClawbackContinuousVestingAccount", vestingAccAddr.String())
	}
	if vestingAccount.GetVestingCoins(ctx.BlockTime()).IsZero() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "account %s does not have currently vesting coins", msg.AccountAddress)
	}
	if vestingAccount.FunderAddress != msg.FunderAddress {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "clawback can only be requested by original funder: %s", vestingAccount.FunderAddress)
	}

	//
	//
	// clawback from agoric x/auth/vesting
	//
	//
	//

	//
	// MODIFYING THIS PART
	//
	// Compute the clawback based on the account state only, and update account
	// toClawBack := vestingAccount.computeClawback(ctx.BlockTime().Unix())
	// if toClawBack.IsZero() {
	// 	return nil
	// }

	totalVesting := vestingAccount.GetVestingCoins(ctx.BlockTime())

	// update the vesting account
	totalVested := vestingAccount.GetVestedCoins(ctx.BlockTime())
	vestingAccount.OriginalVesting = totalVested
	vestingAccount.EndTime = ctx.BlockTime().Unix() // so that all of original vesting is vested now

	//
	//
	//

	// addr := vestingAccount.GetAddress()
	bondDenom, err := m.sk.BondDenom(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bond denomination")
	}

	// Compute the clawback based on bank balance and delegation, and update account
	// encumbered := vestingAccount.GetVestingCoins(ctx.BlockTime())
	encumbered := totalVesting // TO-DO ?????
	bondedAmt, err := m.sk.GetDelegatorBonded(ctx, vestingAccAddr)
	if err != nil {
		return nil, fmt.Errorf("error while getting bonded amount: %w", err)
	}
	unbondingAmt, err := m.sk.GetDelegatorUnbonding(ctx, vestingAccAddr)
	if err != nil {
		return nil, fmt.Errorf("error while getting unbonding amount: %w", err)
	}
	bonded := sdk.NewCoins(sdk.NewCoin(bondDenom, bondedAmt))
	unbonding := sdk.NewCoins(sdk.NewCoin(bondDenom, unbondingAmt))
	unbonded := m.bk.GetAllBalances(ctx, vestingAccAddr)

	//
	// MODIFYING THIS PART
	//
	// toClawBack = vestingAccount.updateDelegation(encumbered, toClawBack, bonded, unbonding, unbonded)
	delegated := bonded.Add(unbonding...)
	oldDelegated := vestingAccount.DelegatedVesting.Add(vestingAccount.DelegatedFree...)
	slashed := oldDelegated.Sub(coinsMin(delegated, oldDelegated)...)
	total := delegated.Add(unbonded...)
	toClawBack := coinsMin(totalVesting, total) // might have been slashed
	newDelegated := coinsMin(delegated, total.Sub(totalVesting...)).Add(slashed...)
	vestingAccount.DelegatedVesting = coinsMin(encumbered, newDelegated)
	vestingAccount.DelegatedFree = newDelegated.Sub(vestingAccount.DelegatedVesting...)

	// Write now now so that the bank module sees unvested tokens are unlocked.
	// Note that all store writes are aborted if there is a panic, so there is
	// no danger in writing incomplete results.
	m.ak.SetAccount(ctx, vestingAccount)

	//
	//
	//

	// Now that future vesting events (and associated lockup) are removed,
	// the balance of the account is unlocked and can be freely transferred.
	spendable := m.bk.SpendableCoins(ctx, vestingAccAddr)
	toXfer := coinsMin(toClawBack, spendable)
	err = m.bk.SendCoins(ctx, vestingAccAddr, funderAddr, toXfer)
	if err != nil {
		return nil, err // shouldn't happen, given spendable check
	}
	toClawBack = toClawBack.Sub(toXfer...)

	// We need to traverse the staking data structures to update the
	// vesting account bookkeeping, and to recover more funds if necessary.
	// Staking is the only way unvested tokens should be missing from the bank balance.

	// If we need more, transfer UnbondingDelegations.
	want := toClawBack.AmountOf(bondDenom)
	unbondings, err := m.sk.GetUnbondingDelegations(ctx, vestingAccAddr, math.MaxUint16)
	if err != nil {
		return nil, fmt.Errorf("error while getting unbonding delegations: %w", err)
	}
	for _, unbonding := range unbondings {
		valAddr, err := sdk.ValAddressFromBech32(unbonding.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		transferred := m.sk.TransferUnbonding(ctx, vestingAccAddr, funderAddr, valAddr, want)
		want = want.Sub(transferred)
		if !want.IsPositive() {
			break
		}
	}

	// If we need more, transfer Delegations.
	if want.IsPositive() {
		delegations, err := m.sk.GetDelegatorDelegations(ctx, vestingAccAddr, math.MaxUint16)
		if err != nil {
			return nil, fmt.Errorf("error while getting delegations: %w", err)
		}

		for _, delegation := range delegations {
			validatorAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
			if err != nil {
				panic(err) // shouldn't happen
			}
			validator, err := m.sk.GetValidator(ctx, validatorAddr)
			if err != nil {
				// validator has been removed
				// TO-DO more specific check?
				continue
			}
			wantShares, err := validator.SharesFromTokensTruncated(want)
			if err != nil {
				// validator has no tokens
				continue
			}
			transferredShares, err := m.sk.TransferDelegation(ctx, vestingAccAddr, funderAddr, validatorAddr, wantShares)
			if err != nil {
				panic(err) // shouldn't happen TO-DO

			}
			// to be conservative in what we're clawing back, round transferred shares up
			transferred := validator.TokensFromSharesRoundUp(transferredShares).RoundInt()
			want = want.Sub(transferred)
			if !want.IsPositive() {
				// Could be slightly negative, due to rounding?
				// Don't think so, due to the precautions above.
				break
			}
		}
	}

	// If we've transferred everything and still haven't transferred the desired clawback amount,
	// then the account must have most some unvested tokens from slashing.

	//
	//
	//
	//
	//
	//

	// // Compute clawback amount, unlock unvested tokens and remove future vesting events
	// // updatedAcc, toClawBack := vestingAccount.ComputeClawback(ctx.BlockTime().Unix())

	// // totalVested := vestingAccount.GetVestedOnly(time.Unix(clawbackTime, 0))
	// // totalUnvested := vestingAccount.GetUnvestedOnly(time.Unix(clawbackTime, 0))

	// totalVested := vestingAccount.GetVestedCoins(ctx.BlockTime())    // consider transferred??
	// totalUnvested := vestingAccount.GetVestingCoins(ctx.BlockTime()) // clawback amount

	// // create new vesting account to overwrite
	// // TO-DO: or just create a non-vesting account
	// newVestingAcc := new(types.ClawbackContinuousVestingAccount)
	// var err error
	// newVestingAcc.ContinuousVestingAccount, err = sdkvestingtypes.NewContinuousVestingAccount(vestingAccount.BaseAccount, totalVested, vestingAccount.StartTime, vestingAccount.EndTime)
	// if err != nil {
	// 	return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "failed to create continuous vesting account: %s")
	// }
	// newVestingAcc.OriginalVesting = totalVested
	// newVestingAcc.EndTime = ctx.BlockTime().Unix() // so that all of original vesting is vested now
	// newVestingAcc.BaseAccount = vestingAccount.BaseAccount

	// m.AccountKeeper.SetAccount(ctx, newVestingAcc)

	// // // Remove all unvested periods from the schedule
	// // passedPeriodID := vestingAccount.GetPassedPeriodCount(time.Unix(clawbackTime, 0))
	// // newVestingPeriods := vestingAccount.VestingPeriods[:passedPeriodID]
	// // newVestingEnd := vestingAccount.GetStartTime() + newVestingPeriods.TotalLength()

	// // Cap the unlocking schedule to the new total vested.
	// //  - If lockup has already passed, all vested coins are unlocked.
	// //  - If lockup has not passed, the vested coins, are still locked.
	// // capPeriods := sdkvesting.Periods{
	// // 	{
	// // 		Length: 0,
	// // 		Amount: totalVested,
	// // 	},
	// // }

	// // minimum of the 2 periods
	// // _, newLockingEnd, newLockupPeriods := ConjunctPeriods(vestingAccount.GetStartTime(), vestingAccount.GetStartTime(), vestingAccount.LockupPeriods, capPeriods)

	// // Now construct the new account state
	// // vestingAccount.EndTime = Max64(newVestingEnd, newLockingEnd)
	// // vestingAccount.LockupPeriods = newLockupPeriods
	// // vestingAccount.VestingPeriods = newVestingPeriods

	// // return vestingAccount, totalUnvested

	// /*
	// 	// convert the account back to a normal EthAccount
	// 	//
	// 	// NOTE: this is necessary to allow the bank keeper to send the locked coins away to the
	// 	// destination address. If the account is not converted, the coins will still be seen as locked,
	// 	// and can therefore not be transferred.
	// 	ethAccount := evmostypes.ProtoAccount().(*evmostypes.EthAccount)
	// 	ethAccount.BaseAccount = updatedAcc.BaseAccount

	// 	// set the account with the updated values of the vesting schedule
	// 	k.accountKeeper.SetAccount(ctx, ethAccount)

	// 	address := updatedAcc.GetAddress()

	// 	// NOTE: don't use `SpendableCoins` to get the minimum value to clawback since
	// 	// the amount is retrieved from `ComputeClawback`, which ensures correctness.
	// 	// `SpendableCoins` can result in gas exhaustion if the user has too many
	// 	// different denoms (because of store iteration).

	// 	// Transfer clawback to the destination (funder)
	// 	// return toClawBack, k.bankKeeper.SendCoins(ctx, address, destinationAddr, toClawBack)

	// 	// clawedBack, err
	// */

	// err = m.BankKeeper.SendCoins(ctx, newVestingAcc.GetAddress(), funderAddress, totalUnvested)
	// if err != nil {
	// 	return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "failed to claw back '%s' from '%s' to '%s'", totalUnvested.String(), accountAddress.String(), funderAddress.String())
	// }

	// /*
	// 	ctx.EventManager().EmitEvents(
	// 		sdk.Events{
	// 			sdk.NewEvent(
	// 				types.EventTypeClawback,
	// 				sdk.NewAttribute(types.AttributeKeyFunder, msg.FunderAddress),
	// 				sdk.NewAttribute(types.AttributeKeyAccount, msg.AccountAddress),
	// 				sdk.NewAttribute(types.AttributeKeyDestination, dest.String()),
	// 			),
	// 		},
	// 	)
	// */

	if !want.IsZero() {
		panic("want not zero")
	}

	return &types.MsgClawbackResponse{
		Coins: toClawBack,
	}, nil
}

func validateAmount(amount sdk.Coins) error {
	if !amount.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrap(amount.String())
	}

	if !amount.IsAllPositive() {
		return sdkerrors.ErrInvalidCoins.Wrap(amount.String())
	}

	return nil
}

// coinsMin returns the minimum of its inputs for all denominations.
func coinsMin(a, b sdk.Coins) sdk.Coins {
	min := sdk.NewCoins()
	for _, coinA := range a {
		denom := coinA.Denom
		bAmt := b.AmountOfNoDenomValidation(denom)
		minAmt := coinA.Amount
		if minAmt.GT(bAmt) {
			minAmt = bAmt
		}
		if minAmt.IsPositive() {
			min = min.Add(sdk.NewCoin(denom, minAmt))
		}
	}
	return min
}
