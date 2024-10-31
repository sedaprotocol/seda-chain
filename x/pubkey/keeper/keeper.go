package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

type Keeper struct {
	stakingKeeper         types.StakingKeeper
	validatorAddressCodec address.Codec

	Schema  collections.Schema
	pubKeys collections.Map[collections.Pair[[]byte, uint32], []byte]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, sk types.StakingKeeper, validatorAddressCodec address.Codec) *Keeper {
	if validatorAddressCodec == nil {
		panic("validator address codec is nil")
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		stakingKeeper:         sk,
		validatorAddressCodec: validatorAddressCodec,
		pubKeys:               collections.NewMap(sb, types.PubKeysPrefix, "pubkeys", collections.PairKeyCodec(collections.BytesKey, collections.Uint32Key), collections.BytesValue),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return &k
}

func (k Keeper) SetValidatorKeyAtIndex(ctx context.Context, validatorAddr sdk.ValAddress, index utils.SEDAKeyIndex, pubKey []byte) error {
	err := k.pubKeys.Set(ctx, collections.Join(validatorAddr.Bytes(), uint32(index)), pubKey)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) GetValidatorKeyAtIndex(ctx context.Context, validatorAddr sdk.ValAddress, index utils.SEDAKeyIndex) ([]byte, error) {
	pubKey, err := k.pubKeys.Get(ctx, collections.Join(validatorAddr.Bytes(), uint32(index)))
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

func (k Keeper) GetValidatorKeys(ctx context.Context, validatorAddr string) (result types.ValidatorPubKeys, err error) {
	valAddr, err := k.validatorAddressCodec.StringToBytes(validatorAddr)
	if err != nil {
		return result, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	rng := collections.NewPrefixedPairRange[[]byte, uint32](valAddr)
	itr, err := k.pubKeys.Iterate(ctx, rng)
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
		result.IndexedPubKeys = append(result.IndexedPubKeys, types.IndexedPubKey{
			Index:  kv.Key.K2(),
			PubKey: kv.Value,
		})
	}
	return result, nil
}
