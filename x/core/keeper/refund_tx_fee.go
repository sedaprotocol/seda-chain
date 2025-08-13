package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// RefundTxFee refunds the fee collected by the DeductFeeDecorator ante handler.
func (k Keeper) RefundTxFee(ctx sdk.Context) error {
	tx, err := k.txDecoder(ctx.TxBytes())
	if err != nil {
		return err
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
