package keeper

import (
	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k *Keeper) EndBlock(ctx sdk.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			k.Logger(ctx).Error("recovered from panic in data-proxy EndBlock", "err", r)
		}
		if err != nil {
			k.Logger(ctx).Error("error in data-proxy EndBlock", "err", err)
		}
		err = nil
	}()

	err = k.ProcessFeeUpdates(ctx)
	if err != nil {
		return
	}
	return
}

func (k *Keeper) ProcessFeeUpdates(ctx sdk.Context) error {
	blockHeight := ctx.BlockHeight()
	pubkeys, err := k.GetFeeUpdatePubKeys(ctx, blockHeight)
	if err != nil {
		return err
	}

	for _, pubkey := range pubkeys {
		proxyConfig, err := k.DataProxyConfigs.Get(ctx, pubkey)
		if err != nil {
			return err
		}

		proxyConfig.Fee = &proxyConfig.FeeUpdate.NewFee
		proxyConfig.FeeUpdate = nil

		if err := k.DataProxyConfigs.Set(ctx, pubkey, proxyConfig); err != nil {
			return err
		}

		if err := k.FeeUpdateQueue.Remove(ctx, collections.Join(blockHeight, pubkey)); err != nil {
			return err
		}

		// TODO emit events
	}
	return nil
}
