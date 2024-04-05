package simulation

import (
	"math/rand"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/x/vesting/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgCreateVestingAccount = "op_weight_msg_create_vesting_account" //nolint:gosec
	OpWeightMsgClawback             = "op_weight_msg_clawback"               //nolint:gosec

	DefaultWeightMsgCreateVestingAccount = 100
	DefaultWeightMsgClawback             = 75

	// Parameters for future operations of vesting account creation:
	blocksUntilDelegate    = 3  // number of blocks to wait for delegate
	blocksUntilUndelegate  = 3  // number of blocks to wait for undelegate after delegate
	maxNumDelegate         = 10 // max number of delegations
	minBlocksUntilClawback = 10 // minimum number of blocks to wait for clawback
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	_ codec.JSONCodec,
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
) simulation.WeightedOperations {
	var weightMsgCreateVestingAccount int
	var weightMsgClawback int

	appParams.GetOrGenerate(OpWeightMsgCreateVestingAccount, &weightMsgCreateVestingAccount, nil, func(_ *rand.Rand) {
		weightMsgCreateVestingAccount = DefaultWeightMsgCreateVestingAccount
	})
	appParams.GetOrGenerate(OpWeightMsgClawback, &weightMsgClawback, nil, func(_ *rand.Rand) {
		weightMsgClawback = DefaultWeightMsgClawback
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateVestingAccount,
			SimulateMsgCreateVestingAccount(txGen, ak, bk, sk),
		),
		simulation.NewWeightedOperation(
			weightMsgClawback,
			SimulateMsgClawback(txGen, ak, bk),
		),
	}
}

// SimulateMsgCreateVestingAccount generates a MsgCreateVestingAccount with random values.
func SimulateMsgCreateVestingAccount(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgCreateVestingAccount{})

		funder, _ := simtypes.RandomAcc(r, accs)
		funderAcc := ak.GetAccount(ctx, funder.Address)

		spendableCoins := bk.SpendableCoins(ctx, funderAcc.GetAddress())
		if err := bk.IsSendEnabledCoins(ctx, spendableCoins...); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, err.Error()), nil, nil
		}
		sendCoins := simtypes.RandSubsetCoins(r, spendableCoins)
		if sendCoins.Empty() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "empty coins slice"), nil, nil
		}
		spendableCoins, hasNeg := spendableCoins.SafeSub(sendCoins...)
		var fees sdk.Coins
		var err error
		if !hasNeg {
			fees, err = simtypes.RandomFees(r, ctx, spendableCoins)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to generate fees"), nil, err
			}
		}

		recipient := simtypes.RandomAccounts(r, 1)[0]

		// Randomly choose vesting durations (minimum 10).
		durationBlocks := r.Intn(10) + minBlocksUntilClawback
		durationTime := 100000*durationBlocks + r.Intn(75000*10)

		// Create and deliver tx.
		msg := types.NewMsgCreateVestingAccount(
			funderAcc.GetAddress(),
			recipient.Address,
			sendCoins,
			ctx.BlockTime().Unix()+int64(durationTime),
			false,
		)
		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{funderAcc.GetAccountNumber()},
			[]uint64{funderAcc.GetSequence()},
			funder.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to generate mock tx"), nil, err
		}
		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}

		// Add future operations.
		var futureOps []simtypes.FutureOperation

		bondDenom, err := sk.BondDenom(ctx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "bond denom not found"), nil, err
		}
		found, sendCoin := sendCoins.Find(bondDenom)
		if !found {
			panic("bond denom not found in sent coins")
		}
		for i, n := 0, r.Intn(maxNumDelegate); i < n; i++ {
			futureOps = append(futureOps, simtypes.FutureOperation{
				BlockHeight: int(ctx.BlockHeight()) + blocksUntilDelegate,
				Op:          simulateMsgDelegate(txGen, ak, bk, sk, recipient, sendCoin),
			})
		}

		futureOps = append(futureOps, simtypes.FutureOperation{
			BlockHeight: int(ctx.BlockHeight()) + durationBlocks,
			Op:          simulateMsgClawbackFutureOp(txGen, ak, bk, sk, recipient, funder),
		})
		return simtypes.NewOperationMsg(msg, true, ""), futureOps, nil
	}
}

// SimulateMsgCreateVestingAccount generates a MsgCreateVestingAccount with random values.
func SimulateMsgClawback(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgClawback{})

		recipient, _ := simtypes.RandomAcc(r, accs)
		recipientAcc := ak.GetAccount(ctx, recipient.Address)
		vestingAcc, isClawback := recipientAcc.(*types.ClawbackContinuousVestingAccount)
		if !isClawback {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "not a vesting account"), nil, nil
		}
		if vestingAcc.GetVestingCoins(ctx.BlockTime()).IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "vesting account not vesting anymore"), nil, nil
		}

		funder, found := simtypes.FindAccount(accs, sdk.MustAccAddressFromBech32(vestingAcc.FunderAddress))
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "failed to find funder account"), nil, nil
		}
		funderAcc := ak.GetAccount(ctx, funder.Address)
		spendableCoins := bk.SpendableCoins(ctx, funderAcc.GetAddress())
		if err := bk.IsSendEnabledCoins(ctx, spendableCoins...); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, err.Error()), nil, nil
		}

		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to generate fees"), nil, err
		}

		msg := types.NewMsgClawback(
			funderAcc.GetAddress(),
			recipient.Address,
		)
		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{funderAcc.GetAccountNumber()},
			[]uint64{funderAcc.GetSequence()},
			funder.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}
		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

func simulateMsgDelegate(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
	acc simtypes.Account,
	originalVesting sdk.Coin,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&stakingtypes.MsgDelegate{})

		vals, err := sk.GetAllValidators(ctx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to get validators"), nil, err
		}
		if len(vals) < 3 {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "number of validators equal zero"), nil, nil
		}
		val, ok := testutil.RandSliceElem(r, vals[2:]) // Spare first two validators so we don't crash the network.
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to pick a validator"), nil, nil
		}
		if val.InvalidExRate() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "validator's invalid echange rate"), nil, nil
		}

		account := ak.GetAccount(ctx, acc.Address)
		delAmt := sdk.NewCoin(originalVesting.Denom, originalVesting.Amount.Quo(math.NewInt(maxNumDelegate)))
		fees := sdk.NewCoins()

		msg := stakingtypes.NewMsgDelegate(acc.Address.String(), val.GetOperator(), delAmt)

		// Removing sim check since we cannot know the account number in advance
		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}

		var futureOps []simtypes.FutureOperation
		// 1/2 undelegate, 1/4 redelegate, 1/4 no future operation.
		if randNum := r.Intn(2); randNum > 0 {
			futureOps = append(futureOps, simtypes.FutureOperation{
				BlockHeight: int(ctx.BlockHeight()) + blocksUntilUndelegate,
				Op:          simulateMsgUndelegate(txGen, ak, sk, acc, val, delAmt),
			})
		} else if randNum := r.Intn(2); randNum > 0 {
			futureOps = append(futureOps, simtypes.FutureOperation{
				BlockHeight: int(ctx.BlockHeight()) + blocksUntilUndelegate,
				Op:          simulateMsgRedelegate(txGen, ak, sk, acc, val, delAmt),
			})
		}
		return simtypes.NewOperationMsg(msg, true, ""), futureOps, nil
	}
}

func simulateMsgUndelegate(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	sk types.StakingKeeper,
	delegator simtypes.Account,
	validator stakingtypes.Validator,
	delAmt sdk.Coin,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{})

		valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
		if err != nil {
			panic(err)
		}

		// Determine total bonded amount.
		val, err := sk.GetValidator(ctx, valAddr)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to get validators"), nil, err
		}
		delegation, err := sk.GetDelegation(ctx, delegator.Address, valAddr)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "error getting validator delegations"), nil, nil
		}
		totalBond := val.TokensFromShares(delegation.GetShares()).TruncateInt()
		if !totalBond.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "total bond is negative"), nil, nil
		}

		amount := totalBond
		// Undelegate total bonded amount 1/5 of the time.
		if randNum := r.Intn(5); randNum >= 1 {
			amount, err = simtypes.RandPositiveInt(r, totalBond)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to generate positive undelegation amount"), nil, err
			}
		}
		undelAmt := sdk.NewCoin(delAmt.Denom, amount)

		account := ak.GetAccount(ctx, delegator.Address)
		fees := sdk.NewCoins()
		msg := stakingtypes.NewMsgUndelegate(delegator.Address.String(), validator.GetOperator(), undelAmt)
		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			delegator.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}
		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

func simulateMsgRedelegate(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	sk types.StakingKeeper,
	delegator simtypes.Account,
	validator stakingtypes.Validator,
	delAmt sdk.Coin,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{})

		valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
		if err != nil {
			panic(err)
		}
		srcVal, err := sk.GetValidator(ctx, valAddr)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to get validators"), nil, err
		}

		// Determine redelegation destination validator.
		vals, err := sk.GetAllValidators(ctx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to get validators"), nil, err
		}
		if len(vals) < 3 {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "number of validators equal zero"), nil, nil
		}
		dstVal, ok := testutil.RandSliceElem(r, vals[2:]) // Spare first two validators so we don't crash the network.
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to pick a validator"), nil, nil
		}
		if dstVal.GetOperator() == srcVal.GetOperator() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "validator's invalid echange rate"), nil, nil
		}
		if dstVal.InvalidExRate() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "validator's invalid echange rate"), nil, nil
		}
		found, err := sk.HasReceivingRedelegation(ctx, delegator.Address, valAddr)
		if err != nil || found {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "redelegation to destination validator in progress already exists"), nil, nil
		}

		// Determine total bonded amount.
		delegation, err := sk.GetDelegation(ctx, delegator.Address, valAddr)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "error getting validator delegations"), nil, nil
		}
		totalBond := srcVal.TokensFromShares(delegation.GetShares()).TruncateInt()
		if totalBond.LTE(math.OneInt()) {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "total bond is less than equal to one"), nil, nil
		}

		amount := totalBond
		// Undelegate total bonded amount 1/5 of the time.
		if randNum := r.Intn(5); randNum >= 1 {
			amount, err = simtypes.RandPositiveInt(r, totalBond)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to generate positive undelegation amount"), nil, err
			}
		}
		redelAmt := sdk.NewCoin(delAmt.Denom, amount)

		account := ak.GetAccount(ctx, delegator.Address)
		fees := sdk.NewCoins()
		msg := stakingtypes.NewMsgBeginRedelegate(delegator.Address.String(), srcVal.GetOperator(), dstVal.GetOperator(), redelAmt)
		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			delegator.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}
		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

func simulateMsgClawbackFutureOp(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
	recipient simtypes.Account,
	funder simtypes.Account,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgClawback{})

		funder, found := simtypes.FindAccount(accs, funder.Address)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "failed to find funder account"), nil, nil
		}
		funderAcc := ak.GetAccount(ctx, funder.Address)
		spendableCoins := bk.SpendableCoins(ctx, funderAcc.GetAddress())
		if err := bk.IsSendEnabledCoins(ctx, spendableCoins...); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, err.Error()), nil, nil
		}

		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to generate fees"), nil, err
		}

		msg := types.NewMsgClawback(
			funderAcc.GetAddress(),
			recipient.Address,
		)

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{funderAcc.GetAccountNumber()},
			[]uint64{funderAcc.GetSequence()},
			funder.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}
		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}
