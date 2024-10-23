package keeper

import (
	"encoding/hex"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) (err error) {
	// Use defer to prevent returning an error, which would cause
	// the chain to halt.
	defer func() {
		// Handle a panic.
		if r := recover(); r != nil {
			k.Logger(ctx).Error("recovered from panic in wasm-storage EndBlock", "err", r)
		}
		// Handle an error.
		if err != nil {
			k.Logger(ctx).Error("error in wasm-storage EndBlock", "err", err)
		}
		err = nil
	}()

	err = k.ExpireOraclePrograms(ctx)
	if err != nil {
		return
	}
	return
}

func (k Keeper) ExpireOraclePrograms(ctx sdk.Context) error {
	blockHeight := ctx.BlockHeight()
	keys, err := k.GetExpiredOracleProgamKeys(ctx, blockHeight)
	if err != nil {
		return err
	}
	for _, hash := range keys {
		if err := k.OracleProgram.Remove(ctx, hash); err != nil {
			return err
		}
		if err := k.OracleProgramExpiration.Remove(ctx, collections.Join(blockHeight, hash)); err != nil {
			return err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeOracleProgramExpiration,
				sdk.NewAttribute(types.AttributeOracleProgramHash, hex.EncodeToString(hash)),
			),
		)
	}
	return nil
}
