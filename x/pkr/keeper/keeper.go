package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

type Keeper struct {
	stakingKeeper         types.StakingKeeper
	validatorAddressCodec address.Codec

	Schema  collections.Schema
	PubKeys collections.Map[collections.Pair[[]byte, uint32], cryptotypes.PubKey]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, sk types.StakingKeeper, validatorAddressCodec address.Codec) *Keeper {
	if validatorAddressCodec == nil {
		panic("validator address codec is nil")
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		stakingKeeper:         sk,
		validatorAddressCodec: validatorAddressCodec,
		PubKeys:               collections.NewMap(sb, types.PubKeysPrefix, "pubkeys", collections.PairKeyCodec(collections.BytesKey, collections.Uint32Key), codec.CollInterfaceValue[cryptotypes.PubKey](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return &k
}
