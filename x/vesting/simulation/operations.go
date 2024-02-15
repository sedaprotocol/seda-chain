package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"

	// codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
			SimulateMsgClawback(txGen, ak, bk, sk),
		),
	}
}

// SimulateMsgCreateVestingAccount generates a MsgCreateVestingAccount with random values.
func SimulateMsgCreateVestingAccount(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	_ types.StakingKeeper,
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

		msg := types.NewMsgCreateVestingAccount(
			funderAcc.GetAddress(),
			recipient.Address,
			sendCoins,
			ctx.BlockTime().Unix()+1000,
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

		// TO-DO activate future operations
		// // future operations
		var futureOps []simtypes.FutureOperation
		// // recipient stakes
		// op := simulateMsgDelegate(txGen, ak, bk, sk, recipient, sendCoins)
		// futureOps = append(futureOps, simtypes.FutureOperation{
		// 	BlockHeight: int(ctx.BlockHeight()) + 1,
		// 	Op:          op,
		// })

		// // then funder claws back
		// op2 := simulateMsgClawbackFutureOp(txGen, ak, bk, sk, recipient, funder)
		// futureOps = append(futureOps, simtypes.FutureOperation{
		// 	BlockHeight: int(ctx.BlockHeight()) + 1,
		// 	Op:          op2,
		// })

		return simtypes.NewOperationMsg(msg, true, ""), futureOps, nil
	}
}

// SimulateMsgCreateVestingAccount generates a MsgCreateVestingAccount with random values.
func SimulateMsgClawback(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	_ types.StakingKeeper,
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

//nolint:unused
func simulateMsgDelegate(
	_ client.TxConfig,
	_ types.AccountKeeper,
	_ types.BankKeeper,
	sk types.StakingKeeper,
	acc simtypes.Account,
	originalVesting sdk.Coins,
) simtypes.Operation {
	return func(
		r *rand.Rand, _ *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&stakingtypes.MsgDelegate{})
		denom, err := sk.BondDenom(ctx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "bond denom not found"), nil, err
		}

		vals, err := sk.GetAllValidators(ctx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to get validators"), nil, err
		}

		if len(vals) < 3 {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "number of validators equal zero"), nil, nil
		}

		// simAccount, _ := simtypes.RandomAcc(r, accs)

		// TO-DO: spare first two validators so we don't crash the network
		val, ok := testutil.RandSliceElem(r, vals[2:])
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to pick a validator"), nil, nil
		}

		if val.InvalidExRate() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "validator's invalid echange rate"), nil, nil
		}

		// amount := bk.GetBalance(ctx, acc.Address, denom).Amount
		// if !amount.IsPositive() {
		// 	return simtypes.NoOpMsg(types.ModuleName, msgType, "balance is negative"), nil, nil
		// }

		// amount, err = simtypes.RandPositiveInt(r, amount)
		// if err != nil {
		// 	return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to generate positive amount"), nil, err
		// }

		// bondAmt := sdk.NewCoin(denom, amount)

		// account := ak.GetAccount(ctx, acc.Address)
		// spendable := bk.SpendableCoins(ctx, account.GetAddress())

		// fees := sdk.NewCoins()

		// coins, hasNeg := spendable.SafeSub(bondAmt)
		// if !hasNeg {
		// 	fees, err = simtypes.RandomFees(r, ctx, coins)
		// 	if err != nil {
		// 		return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to generate fees"), nil, err
		// 	}
		// }
		found, bondAmt := originalVesting.Find(denom)
		if !found {
			panic("no bond denom in original vesting coins")
		}
		msg := stakingtypes.NewMsgDelegate(acc.Address.String(), val.GetOperator(), bondAmt)

		// Removing sim check since we cannot know the account number in advance
		// tx, err := simtestutil.GenSignedMockTx(
		// 	r,
		// 	txGen,
		// 	[]sdk.Msg{msg},
		// 	fees,
		// 	simtestutil.DefaultGenTxGas,
		// 	chainID,
		// 	[]uint64{uint64(numAccs)}, //[]uint64{account.GetAccountNumber()},
		// 	[]uint64{0},               //[]uint64{account.GetSequence()},
		// 	acc.PrivKey,
		// )
		// if err != nil {
		// 	return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to generate mock tx"), nil, err
		// }

		// _, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		// if err != nil {
		// 	return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		// }
		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

//nolint:unused
func simulateMsgClawbackFutureOp(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	_ types.StakingKeeper,
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

		// fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		// if err != nil {
		// 	return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to generate fees"), nil, err
		// }

		msg := types.NewMsgClawback(
			funderAcc.GetAddress(),
			recipient.Address,
		)

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			sdk.NewCoins(), // fees,
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
