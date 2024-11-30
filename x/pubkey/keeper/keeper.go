package keeper

import (
	"bytes"
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

type Keeper struct {
	stakingKeeper         types.StakingKeeper
	validatorAddressCodec address.Codec

	Schema         collections.Schema
	pubKeys        collections.Map[collections.Pair[[]byte, uint32], []byte]
	provingSchemes collections.Map[uint32, bool]
}

func NewKeeper(storeService storetypes.KVStoreService, sk types.StakingKeeper, validatorAddressCodec address.Codec) *Keeper {
	if validatorAddressCodec == nil {
		panic("validator address codec is nil")
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		stakingKeeper:         sk,
		validatorAddressCodec: validatorAddressCodec,
		pubKeys:               collections.NewMap(sb, types.PubKeysPrefix, "pubkeys", collections.PairKeyCodec(collections.BytesKey, collections.Uint32Key), collections.BytesValue),
		provingSchemes:        collections.NewMap(sb, types.ProvingSchemesPrefix, "proving_schemes", collections.Uint32Key, collections.BoolValue),
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

// GetValidatorKeys returns all public keys of a given validator.
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

// GetAllValidatorPubKeys returns all validator public keys in the store.
func (k Keeper) GetAllValidatorPubKeys(ctx context.Context) ([]types.ValidatorPubKeys, error) {
	var valPubKeys []types.ValidatorPubKeys

	itr, err := k.pubKeys.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer itr.Close()

	var currentVal []byte
	for ; itr.Valid(); itr.Next() {
		kv, err := itr.KeyValue()
		if err != nil {
			return nil, err
		}

		// Skip if the validator has already been processed.
		if bytes.Equal(kv.Key.K1(), currentVal) {
			continue
		}
		currentVal = kv.Key.K1()

		valAddr, err := k.validatorAddressCodec.BytesToString(kv.Key.K1())
		if err != nil {
			return nil, err
		}
		res, err := k.GetValidatorKeys(ctx, valAddr)
		if err != nil {
			return nil, err
		}

		valPubKeys = append(valPubKeys, res)
	}
	return valPubKeys, err
}

func (k Keeper) SetProvingScheme(ctx context.Context, index utils.SEDAKeyIndex, isEnabled bool) error {
	err := k.provingSchemes.Set(ctx, uint32(index), isEnabled)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) EnableProvingScheme(ctx context.Context, index utils.SEDAKeyIndex) error {
	err := k.provingSchemes.Set(ctx, uint32(index), true)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) IsProvingSchemeEnabled(ctx context.Context, index utils.SEDAKeyIndex) (bool, error) {
	isEnabled, err := k.provingSchemes.Get(ctx, uint32(index))
	if err != nil {
		return false, err
	}
	return isEnabled, nil
}

func (k Keeper) GetAllProvingSchemes(ctx sdk.Context) ([]types.ProvingScheme, error) {
	itr, err := k.provingSchemes.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer itr.Close()

	var schemes []types.ProvingScheme
	for ; itr.Valid(); itr.Next() {
		kv, err := itr.KeyValue()
		if err != nil {
			return nil, err
		}

		isEnabled, err := k.provingSchemes.Get(ctx, kv.Key)
		if err != nil {
			return nil, err
		}
		schemes = append(schemes, types.ProvingScheme{
			Index:     kv.Key,
			IsEnabled: isEnabled,
		})
	}
	return schemes, nil
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
