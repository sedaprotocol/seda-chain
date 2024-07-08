package keeper

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

type Keeper struct {
	Modules []string

	Schema     collections.Schema
	PublicKeys collections.Map[collections.Pair[string, string], cryptotypes.PubKey]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		PublicKeys: collections.NewMap(sb, types.VRFKeyPrefix, "vrf_key", collections.PairKeyCodec(collections.StringKey, collections.StringKey), codec.CollInterfaceValue[cryptotypes.PubKey](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema
	return &k
}
