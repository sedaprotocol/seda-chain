package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"

	// codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/sedaprotocol/seda-chain/x/vesting/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgCreateVestingAccount = "op_weight_msg_create_vesting_account"
	OpWeightMsgClawback             = "op_weight_msg_create_vesting_account"

	DefaultWeightMsgCreateVestingAccount = 100
	DefaultWeightMsgClawback             = 100
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	// registry codectypes.InterfaceRegistry,
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreateVestingAccount int
		weightMsgClawback             int
	)
	appParams.GetOrGenerate(OpWeightMsgCreateVestingAccount, &weightMsgCreateVestingAccount, nil, func(_ *rand.Rand) {
		weightMsgCreateVestingAccount = DefaultWeightMsgCreateVestingAccount
	})
	appParams.GetOrGenerate(OpWeightMsgClawback, &weightMsgClawback, nil, func(_ *rand.Rand) {
		weightMsgClawback = DefaultWeightMsgClawback
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateVestingAccount,
			SimulateMsgCreateVestingAccount(
				// codec.NewProtoCodec(registry),
				txGen,
				ak, bk, sk),
		),
		simulation.NewWeightedOperation(
			weightMsgClawback,
			SimulateMsgClawback(
				// codec.NewProtoCodec(registry),
				txGen,
				ak, bk, sk),
		),
	}
}

// SimulateMsgCreateVestingAccount generates a MsgCreateVestingAccount with random values.
func SimulateMsgCreateVestingAccount(
	// cdc *codec.ProtoCodec,
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

		msg := types.NewMsgCreateVestingAccount(
			funderAcc.GetAddress(),
			recipient.Address,
			sendCoins,
			ctx.BlockTime().Unix()+1000,
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

// SimulateMsgCreateVestingAccount generates a MsgCreateVestingAccount with random values.
func SimulateMsgClawback(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
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
