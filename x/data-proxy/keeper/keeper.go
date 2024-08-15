package keeper

import (
	"context"
	"encoding/hex"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

type Keeper struct {
	// authority is the address capable of executing MsgUpdateParams
	// or MsgStoreExecutorWasm. Typically, this should be the gov module
	// address.
	authority string

	Schema           collections.Schema
	DataProxyConfigs collections.Map[[]byte, types.ProxyConfig]
	// TODO queue
	Params collections.Item[types.Params]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, authority string) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		authority:        authority,
		DataProxyConfigs: collections.NewMap(sb, types.DataProxyConfigPrefix, "configs", collections.BytesKey, codec.CollValue[types.ProxyConfig](cdc)),
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
