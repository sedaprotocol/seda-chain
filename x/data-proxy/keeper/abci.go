package keeper

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

func (k *Keeper) EndBlock(ctx sdk.Context) error {
	return k.ProcessFeeUpdates(ctx)
}

func (k *Keeper) ProcessFeeUpdates(ctx sdk.Context) error {
	blockHeight := ctx.BlockHeight()
	pubkeys, err := k.GetFeeUpdatePubKeys(ctx, blockHeight)
	if err != nil {
		return err
	}

	for _, pubkey := range pubkeys {
		proxyConfig, err := k.GetDataProxyConfig(ctx, pubkey)
		if err != nil {
			return err
		}

		proxyConfig.Fee = proxyConfig.FeeUpdate.NewFee
		proxyConfig.FeeUpdate = nil

		if err := k.SetDataProxyConfig(ctx, pubkey, proxyConfig); err != nil {
			return err
		}

		if err := k.RemoveFeeUpdate(ctx, blockHeight, pubkey); err != nil {
			return err
		}

		pubKeyHex := hex.EncodeToString(pubkey)
		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeFeeUpdate,
			sdk.NewAttribute(types.AttributePubKey, pubKeyHex),
			sdk.NewAttribute(types.AttributeFee, proxyConfig.Fee.String())))
	}
	return nil
}
