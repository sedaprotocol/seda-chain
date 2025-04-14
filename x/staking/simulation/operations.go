package simulation

import (
	"math/rand"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	sdkstakingsimulation "github.com/cosmos/cosmos-sdk/x/staking/simulation"
	sdktypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/staking/keeper"
	"github.com/sedaprotocol/seda-chain/x/staking/types"
)

func WeightedOperations(
	appParams simtypes.AppParams,
	_ codec.JSONCodec,
	txGen client.TxConfig,
	ak sdktypes.AccountKeeper,
	bk sdktypes.BankKeeper,
	k *keeper.Keeper,
) simulation.WeightedOperations {
	var (
		// We'll reuse all the settings for the original MsgCreateValidator from the staking module,
		// only replacing the implementation of SimulateMsgCreateValidator with our own.
		weightMsgCreateValidator           int
		weightMsgEditValidator             int
		weightMsgDelegate                  int
		weightMsgUndelegate                int
		weightMsgBeginRedelegate           int
		weightMsgCancelUnbondingDelegation int
	)

	appParams.GetOrGenerate(sdkstakingsimulation.OpWeightMsgCreateValidator, &weightMsgCreateValidator, nil, func(_ *rand.Rand) {
		weightMsgCreateValidator = sdkstakingsimulation.DefaultWeightMsgCreateValidator
	})

	appParams.GetOrGenerate(sdkstakingsimulation.OpWeightMsgEditValidator, &weightMsgEditValidator, nil, func(_ *rand.Rand) {
		weightMsgEditValidator = sdkstakingsimulation.DefaultWeightMsgEditValidator
	})

	appParams.GetOrGenerate(sdkstakingsimulation.OpWeightMsgDelegate, &weightMsgDelegate, nil, func(_ *rand.Rand) {
		weightMsgDelegate = sdkstakingsimulation.DefaultWeightMsgDelegate
	})

	appParams.GetOrGenerate(sdkstakingsimulation.OpWeightMsgUndelegate, &weightMsgUndelegate, nil, func(_ *rand.Rand) {
		weightMsgUndelegate = sdkstakingsimulation.DefaultWeightMsgUndelegate
	})

	appParams.GetOrGenerate(sdkstakingsimulation.OpWeightMsgBeginRedelegate, &weightMsgBeginRedelegate, nil, func(_ *rand.Rand) {
		weightMsgBeginRedelegate = sdkstakingsimulation.DefaultWeightMsgBeginRedelegate
	})

	appParams.GetOrGenerate(sdkstakingsimulation.OpWeightMsgCancelUnbondingDelegation, &weightMsgCancelUnbondingDelegation, nil, func(_ *rand.Rand) {
		weightMsgCancelUnbondingDelegation = sdkstakingsimulation.DefaultWeightMsgCancelUnbondingDelegation
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateValidator,
			SimulateMsgCreateSEDAValidator(txGen, ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgEditValidator,
			sdkstakingsimulation.SimulateMsgEditValidator(txGen, ak, bk, k.Keeper),
		),
		simulation.NewWeightedOperation(
			weightMsgDelegate,
			sdkstakingsimulation.SimulateMsgDelegate(txGen, ak, bk, k.Keeper),
		),
		simulation.NewWeightedOperation(
			weightMsgUndelegate,
			sdkstakingsimulation.SimulateMsgUndelegate(txGen, ak, bk, k.Keeper),
		),
		simulation.NewWeightedOperation(
			weightMsgBeginRedelegate,
			sdkstakingsimulation.SimulateMsgBeginRedelegate(txGen, ak, bk, k.Keeper),
		),
		simulation.NewWeightedOperation(
			weightMsgCancelUnbondingDelegation,
			sdkstakingsimulation.SimulateMsgCancelUnbondingDelegate(txGen, ak, bk, k.Keeper),
		),
	}
}

// SimulateMsgCreateSEDAValidator generates a MsgCreateSEDAValidator with random values.
// Mostly copied from the original SimulateMsgCreateValidator from the staking module,
func SimulateMsgCreateSEDAValidator(
	txGen client.TxConfig,
	ak sdktypes.AccountKeeper,
	bk sdktypes.BankKeeper,
	k *keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgCreateSEDAValidator{})

		simAccount, _ := simtypes.RandomAcc(r, accs)
		address := sdk.ValAddress(simAccount.Address)

		// ensure the validator doesn't exist already
		_, err := k.GetValidator(ctx, address)
		if err == nil {
			return simtypes.NoOpMsg(sdktypes.ModuleName, msgType, "validator already exists"), nil, nil
		}

		denom, err := k.BondDenom(ctx)
		if err != nil {
			return simtypes.NoOpMsg(sdktypes.ModuleName, msgType, "bond denom not found"), nil, err
		}

		balance := bk.GetBalance(ctx, simAccount.Address, denom).Amount
		if !balance.IsPositive() {
			return simtypes.NoOpMsg(sdktypes.ModuleName, msgType, "balance is negative"), nil, nil
		}

		amount, err := simtypes.RandPositiveInt(r, balance)
		if err != nil {
			return simtypes.NoOpMsg(sdktypes.ModuleName, msgType, "unable to generate positive amount"), nil, err
		}

		selfDelegation := sdk.NewCoin(denom, amount)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		var fees sdk.Coins

		coins, hasNeg := spendable.SafeSub(selfDelegation)
		if !hasNeg {
			fees, err = simtypes.RandomFees(r, ctx, coins)
			if err != nil {
				return simtypes.NoOpMsg(sdktypes.ModuleName, msgType, "unable to generate fees"), nil, err
			}
		}

		description := sdktypes.NewDescription(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
		)

		maxCommission := math.LegacyNewDecWithPrec(int64(simtypes.RandIntBetween(r, 0, 100)), 2)
		commission := sdktypes.NewCommissionRates(
			simtypes.RandomDecAmount(r, maxCommission),
			maxCommission,
			simtypes.RandomDecAmount(r, maxCommission),
		)

		sedaPubKeys, err := utils.GenerateSEDAKeys(address, "seda_keys.json", "", true)
		if err != nil {
			return simtypes.NoOpMsg(sdktypes.ModuleName, msgType, "unable to generate SEDA keys"), nil, err
		}

		msg, err := types.NewMsgCreateSEDAValidator(address.String(), simAccount.ConsKey.PubKey(), sedaPubKeys, selfDelegation, description, commission, math.OneInt())
		if err != nil {
			return simtypes.NoOpMsg(sdktypes.ModuleName, sdk.MsgTypeURL(msg), "unable to create CreateValidator message"), nil, err
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         txGen,
			Cdc:           nil,
			Msg:           msg,
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			ModuleName:    sdktypes.ModuleName,
		}

		return simulation.GenAndDeliverTx(txCtx, fees)
	}
}
