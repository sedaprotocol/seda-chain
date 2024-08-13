package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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

func (k Keeper) GetValidatorKeys(ctx context.Context, validatorAddr string) (result types.ValidatorPubKeys, err error) {
	valAddr, err := k.validatorAddressCodec.StringToBytes(validatorAddr)
	if err != nil {
		return result, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	rng := collections.NewPrefixedPairRange[[]byte, uint32](valAddr)
	itr, err := k.PubKeys.Iterate(ctx, rng)
	if err != nil {
		return result, err
	}
	defer itr.Close()

	kvs, err := itr.KeyValues()
	if err != nil {
		return result, err
	}
	if len(kvs) == 0 {
		return result, sdkerrors.ErrNotFound
	}

	result.ValidatorAddr = validatorAddr
	for _, kv := range kvs {
		pkAny, err := codectypes.NewAnyWithValue(kv.Value)
		if err != nil {
			panic(err)
		}
		result.IndexedPubKeys = append(result.IndexedPubKeys, types.IndexedPubKey{
			Index:  kv.Key.K2(),
			PubKey: pkAny,
		})
	}
	return result, nil
}
