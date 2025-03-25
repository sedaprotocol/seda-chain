package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

type Keeper struct {
	bankKeeper types.BankKeeper

	// authority is the address capable of executing MsgUpdateParams. Typically, this should be the gov module address.
	authority string

	Schema           collections.Schema
	dataProxyConfigs collections.Map[[]byte, types.ProxyConfig]
	feeUpdateQueue   collections.KeySet[collections.Pair[int64, []byte]]
	params           collections.Item[types.Params]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, bk types.BankKeeper, authority string) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		authority:        authority,
		bankKeeper:       bk,
		dataProxyConfigs: collections.NewMap(sb, types.DataProxyConfigPrefix, "configs", collections.BytesKey, codec.CollValue[types.ProxyConfig](cdc)),
		feeUpdateQueue:   collections.NewKeySet(sb, types.FeeUpdatesPrefix, "fee_updates", collections.PairKeyCodec(collections.Int64Key, collections.BytesKey)),
		params:           collections.NewItem(sb, types.ParamsPrefix, "params", codec.CollValue[types.Params](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return &k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) HasDataProxy(ctx sdk.Context, pubKey []byte) (bool, error) {
	return k.dataProxyConfigs.Has(ctx, pubKey)
}

func (k Keeper) SetDataProxyConfig(ctx context.Context, pubKey []byte, proxyConfig types.ProxyConfig) error {
	return k.dataProxyConfigs.Set(ctx, pubKey, proxyConfig)
}

func (k Keeper) GetDataProxyConfig(ctx context.Context, pubKey []byte) (result types.ProxyConfig, err error) {
	config, err := k.dataProxyConfigs.Get(ctx, pubKey)
	if err != nil {
		return types.ProxyConfig{}, err
	}

	return config, nil
}

func (k Keeper) SetFeeUpdate(ctx sdk.Context, height int64, pubKey []byte) error {
	return k.feeUpdateQueue.Set(ctx, collections.Join(height, pubKey))
}

func (k Keeper) RemoveFeeUpdate(ctx sdk.Context, height int64, pubKey []byte) error {
	return k.feeUpdateQueue.Remove(ctx, collections.Join(height, pubKey))
}

func (k Keeper) scheduleFeeUpdate(ctx sdk.Context, pubKeyBytes []byte, proxyConfig types.ProxyConfig, newFee *sdk.Coin, updateDelay uint32) (int64, error) {
	// Determine update height
	updateHeight := ctx.BlockHeight() + int64(updateDelay)
	feeUpdate := &types.FeeUpdate{
		NewFee:       newFee,
		UpdateHeight: updateHeight,
	}

	// Check if the max updates per block is reached
	keys, err := k.getFeeUpdateKeys(ctx, updateHeight)
	if err != nil {
		return 0, err
	}

	if len(keys) >= types.MaxUpdatesPerBlock {
		return 0, types.ErrMaxUpdatesReached
	}

	// Delete previous pending update, if applicable
	if proxyConfig.FeeUpdate != nil {
		err := k.RemoveFeeUpdate(ctx, proxyConfig.FeeUpdate.UpdateHeight, pubKeyBytes)
		if err != nil {
			return 0, err
		}
	}

	// Schedule new update
	proxyConfig.FeeUpdate = feeUpdate
	err = k.SetFeeUpdate(ctx, updateHeight, pubKeyBytes)
	if err != nil {
		return 0, err
	}

	err = k.SetDataProxyConfig(ctx, pubKeyBytes, proxyConfig)
	if err != nil {
		return 0, err
	}

	return updateHeight, nil
}

func (k Keeper) getFeeUpdateKeys(ctx sdk.Context, height int64) ([]collections.Pair[int64, []byte], error) {
	itr, err := k.feeUpdateQueue.Iterate(ctx, collections.NewPrefixedPairRange[int64, []byte](height))
	if err != nil {
		return nil, err
	}

	return itr.Keys()
}

func (k Keeper) GetFeeUpdatePubKeys(ctx sdk.Context, activationHeight int64) ([][]byte, error) {
	pubkeys := make([][]byte, 0)

	keys, err := k.getFeeUpdateKeys(ctx, activationHeight)
	if err != nil {
		return nil, err
	}

	for _, k := range keys {
		pubkeys = append(pubkeys, k.K2())
	}

	return pubkeys, nil
}

func (k Keeper) HasFeeUpdate(ctx sdk.Context, height int64, pubKey []byte) (bool, error) {
	return k.feeUpdateQueue.Has(ctx, collections.Join(height, pubKey))
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
