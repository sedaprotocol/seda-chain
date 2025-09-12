package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// RefundTxFees refunds the fee collected by the DeductFeeDecorator ante handler.
func (k Keeper) RefundTxFees(ctx sdk.Context, drID, publicKey string, isReveal bool) error {
	tx, err := k.txDecoder(ctx.TxBytes())
	if err != nil {
		return err
	}

	// Proceed only if the current message is the last message in the tx.
	// We identify the last message in the tx based on the exclusivity and
	// uniqueness of commit/reveal txs guaranteed by CommitRevealDecorator
	// ante handler.
	msgs := tx.GetMsgs()
	coreContract, err := k.wasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil {
		return err
	}
	msgInfo, ok := types.ExtractCommitRevealMsgInfo(coreContract.String(), msgs[len(msgs)-1])
	if !ok {
		return sdkerrors.ErrLogic.Wrap("must be commit or reveal message")
	}
	if msgInfo != fmt.Sprintf("%s,%s,%t", drID, publicKey, isReveal) {
		return nil
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return sdkerrors.ErrInvalidRequest.Wrap("transaction is not a fee tx")
	}
	// Fee may be zero in simulation setting.
	if feeTx.GetFee().IsZero() {
		return nil
	}

	// Refund the fee if the gas limit is set within a reasonable range.
	if ctx.GasMeter().Limit() < ctx.GasMeter().GasConsumed()*3 {
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, k.feeCollectorName, feeTx.FeePayer(), feeTx.GetFee())
		if err != nil {
			return err
		}
	} else {
		// Fee is not returned when the gas limit is too large to prevent
		// attacks to exhaust the block gas limit.
		ctx.Logger().Info("gas limit too large for refund", "gas_limit", ctx.GasMeter().Limit(), "gas_consumed", ctx.GasMeter().GasConsumed())
	}
	return nil
}
