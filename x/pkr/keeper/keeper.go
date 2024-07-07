package keeper

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

type Keeper struct {
	Schema  collections.Schema
	KeyName collections.Item[string]
}

func NewKeeper(storeService storetypes.KVStoreService) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		KeyName: collections.NewItem(sb, types.VRFKeyPrefix, "vrf_key_id", collections.StringValue),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema
	return &k
}
