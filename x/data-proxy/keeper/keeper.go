package keeper

import (
	"context"
	"encoding/hex"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

type Keeper struct {
	// authority is the address capable of executing MsgUpdateParams
	// or MsgStoreExecutorWasm. Typically, this should be the gov module
	// address.
	authority string

	Schema           collections.Schema
	DataProxyConfigs collections.Map[[]byte, types.ProxyConfig]
	FeeUpdateQueue   collections.KeySet[collections.Pair[int64, []byte]]
	Params           collections.Item[types.Params]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, authority string) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		authority:        authority,
		DataProxyConfigs: collections.NewMap(sb, types.DataProxyConfigPrefix, "configs", collections.BytesKey, codec.CollValue[types.ProxyConfig](cdc)),
		FeeUpdateQueue:   collections.NewKeySet(sb, types.FeeUpdatesPrefix, "fee_updates", collections.PairKeyCodec(collections.Int64Key, collections.BytesKey)),
		Params:           collections.NewItem(sb, types.ParamsPrefix, "params", codec.CollValue[types.Params](cdc)),
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

func (k Keeper) GetDataProxyConfig(ctx context.Context, pubKey string) (result types.ProxyConfig, err error) {
	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return types.ProxyConfig{}, err
	}

	config, err := k.DataProxyConfigs.Get(ctx, pubKeyBytes)
	if err != nil {
		return types.ProxyConfig{}, err
	}

	return config, nil
}

func (k Keeper) GetFeeUpdatePubKeys(ctx context.Context, activationHeight int64) ([][]byte, error) {
	pubkeys := make([][]byte, 0)
	rng := collections.NewPrefixedPairRange[int64, []byte](activationHeight)

	itr, err := k.FeeUpdateQueue.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}

	keys, err := itr.Keys()
	if err != nil {
		return nil, err
	}

	for _, k := range keys {
		pubkeys = append(pubkeys, k.K2())
	}

	return pubkeys, nil
}

func (k Keeper) processProxyFeeUpdate(ctx sdk.Context, pubKeyBytes []byte, proxyConfig *types.ProxyConfig, newFee *sdk.Coin, updateDelay uint32) (int64, error) {

	// Determine update height
	updateHeight := ctx.BlockHeight() + int64(updateDelay)
	feeUpdate := &types.FeeUpdate{
		NewFee:       *newFee,
		UpdateHeight: updateHeight,
	}

	// Delete previous pending update, if applicable
	if proxyConfig.FeeUpdate != nil {
		err := k.FeeUpdateQueue.Remove(ctx, collections.Join(proxyConfig.FeeUpdate.UpdateHeight, pubKeyBytes))
		if err != nil {
			return 0, err
		}
	}

	// Schedule new update
	proxyConfig.FeeUpdate = feeUpdate
	err := k.FeeUpdateQueue.Set(ctx, collections.Join(updateHeight, pubKeyBytes))
	if err != nil {
		return 0, err
	}

	err = k.DataProxyConfigs.Set(ctx, pubKeyBytes, *proxyConfig)
	if err != nil {
		return 0, err
	}

	return updateHeight, nil
}
