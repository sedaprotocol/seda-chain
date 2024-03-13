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
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

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
	if msg.DisableClawback {
		vestingAccount.FunderAddress = ""
	}

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

// Clawback returns the vesting amount from a ClawbackVestingAccount to the funder.
// The funds are transferred from the following sources
// 1. vesting funds that have not been used towards delegation
// 2. delegations
// 3. unbonding delegations
// in this order, as much as possible, until the vesting amount is met. Note that
// due to slashing the funds from all three of these sources may still fail to meet
// the vesting amount according to the vesting schedule.
func (m msgServer) Clawback(goCtx context.Context, msg *types.MsgClawback) (*types.MsgClawbackResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	vestingAccAddr := sdk.MustAccAddressFromBech32(msg.AccountAddress)
	funderAddr := sdk.MustAccAddressFromBech32(msg.FunderAddress)

	// NOTE: we check the destination address only for the case where it's not sent from the
	// authority account, because in that case the destination address is hardcored to the
	// community pool address anyway (see further below).
	if m.bk.BlockedAddr(funderAddr) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized,
			"%s is a blocked address and not allowed to receive funds", msg.FunderAddress,
		)
	}

	// retrieve the vesting account and perform preliminary checks
	acc := m.ak.GetAccount(ctx, vestingAccAddr)
	if acc == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "account at address '%s' does not exist", vestingAccAddr.String())
	}
	vestingAccount, isClawback := acc.(*types.ClawbackContinuousVestingAccount)
	if !isClawback {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "account %s is not of type ClawbackContinuousVestingAccount", vestingAccAddr.String())
	}
	if vestingAccount.GetVestingCoins(ctx.BlockTime()).IsZero() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "account %s does not have currently vesting coins", msg.AccountAddress)
	}
	if vestingAccount.FunderAddress == "" {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "vesting account has no funder registered (clawback disabled): %s", msg.AccountAddress)
	}
	if vestingAccount.FunderAddress != msg.FunderAddress {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "clawback can only be requested by original funder: %s", vestingAccount.FunderAddress)
	}

	// Compute the clawback based on bank balance and delegation, and update account
	bondedAmt, err := m.sk.GetDelegatorBonded(ctx, vestingAccAddr)
	if err != nil {
		return nil, fmt.Errorf("error while getting bonded amount: %w", err)
	}
	unbondingAmt, err := m.sk.GetDelegatorUnbonding(ctx, vestingAccAddr)
	if err != nil {
		return nil, fmt.Errorf("error while getting unbonding amount: %w", err)
	}
	bondDenom, err := m.sk.BondDenom(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bond denomination")
	}

	bonded := sdk.NewCoins(sdk.NewCoin(bondDenom, bondedAmt))
	unbonding := sdk.NewCoins(sdk.NewCoin(bondDenom, unbondingAmt))
	unbonded := m.bk.GetAllBalances(ctx, vestingAccAddr)
	total := bonded.Add(unbonding...).Add(unbonded...)

	toClawBack := coinsMin(vestingAccount.GetVestingCoins(ctx.BlockTime()), total) // might have been slashed

	// Write now now so that the bank module sees unvested tokens are unlocked.
	// Note that all store writes are aborted if there is a panic, so there is
	// no danger in writing incomplete results.
	vestingAccount.EndTime = ctx.BlockTime().Unix() // so that all of original vesting is vested now
	m.ak.SetAccount(ctx, vestingAccount)

	// Now that future vesting events (and associated lockup) are removed,
	// the balance of the account is unlocked and can be freely transferred.
	spendable := m.bk.SpendableCoins(ctx, vestingAccAddr)
	toXfer := coinsMin(toClawBack, spendable)
	if toXfer.IsAllPositive() {
		err = m.bk.SendCoins(ctx, vestingAccAddr, funderAddr, toXfer)
		if err != nil {
			return nil, err // shouldn't happen, given spendable check
		}
	}

	clawedBackUnbonded := toXfer
	clawedBackUnbonding := sdk.NewCoins()
	clawedBackBonded := sdk.NewCoins()
	toClawBack = toClawBack.Sub(toXfer...)
	if !toClawBack.IsZero() {
		// claw back from staking (unbonding delegations then bonded delegations)
		toClawBackStaking := toClawBack.AmountOf(bondDenom)

		// first from unbonding delegations
		unbondings, err := m.sk.GetUnbondingDelegations(ctx, vestingAccAddr, math.MaxUint16)
		if err != nil {
			return nil, fmt.Errorf("error while getting unbonding delegations: %w", err)
		}
		for _, unbonding := range unbondings {
			valAddr, err := sdk.ValAddressFromBech32(unbonding.ValidatorAddress)
			if err != nil {
				return nil, err
			}

			transferred, err := m.sk.TransferUnbonding(ctx, vestingAccAddr, funderAddr, valAddr, toClawBackStaking)
			if err != nil {
				return nil, err
			}

			clawedBackUnbonding = clawedBackUnbonding.Add(sdk.NewCoin(bondDenom, transferred))
			toClawBackStaking = toClawBackStaking.Sub(transferred)
			if !toClawBackStaking.IsPositive() {
				break
			}
		}

		// then from bonded delegations
		if toClawBackStaking.IsPositive() {
			delegations, err := m.sk.GetDelegatorDelegations(ctx, vestingAccAddr, math.MaxUint16)
			if err != nil {
				return nil, fmt.Errorf("error while getting delegations: %w", err)
			}

			for _, delegation := range delegations {
				validatorAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
				if err != nil {
					return nil, err
				}
				validator, err := m.sk.GetValidator(ctx, validatorAddr)
				if err != nil {
					if err != stakingtypes.ErrNoValidatorFound {
						return nil, err
					}
					// validator has been removed
					continue
				}
				wantShares, err := validator.SharesFromTokensTruncated(toClawBackStaking)
				if err != nil {
					// validator has no tokens
					continue
				}

				transferredShares, err := m.sk.TransferDelegation(ctx, vestingAccAddr, funderAddr, validatorAddr, wantShares)
				if err != nil {
					return nil, err
				}

				// to be conservative in what we're clawing back, round transferred shares up
				transferred := validator.TokensFromSharesRoundUp(transferredShares).RoundInt()
				clawedBackBonded = clawedBackBonded.Add(sdk.NewCoin(bondDenom, transferred))
				toClawBackStaking = toClawBackStaking.Sub(transferred)
				if !toClawBackStaking.IsPositive() {
					// Could be slightly negative, due to rounding?
					// Don't think so, due to the precautions above.
					break
				}
			}
		}

		if !toClawBackStaking.IsZero() {
			return nil, fmt.Errorf("failed to claw back full amount")
		}
	}

	return &types.MsgClawbackResponse{
		ClawedUnbonded:  clawedBackUnbonded,
		ClawedUnbonding: clawedBackUnbonding,
		ClawedBonded:    clawedBackBonded,
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
