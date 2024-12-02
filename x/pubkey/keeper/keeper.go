package keeper

import (
	"bytes"
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

// ActivationLag is the number of blocks to wait before activating
// a proving scheme once the threshold of public key registration rate
// is reached.
const ActivationLag = 25

type Keeper struct {
	stakingKeeper         types.StakingKeeper
	validatorAddressCodec address.Codec

	Schema         collections.Schema
	pubKeys        collections.Map[collections.Pair[[]byte, uint32], []byte]
	provingSchemes collections.Map[uint32, types.ProvingScheme]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, sk types.StakingKeeper, valAddrCdc address.Codec) *Keeper {
	if valAddrCdc == nil {
		panic("validator address codec is nil")
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		stakingKeeper:         sk,
		validatorAddressCodec: valAddrCdc,
		pubKeys:               collections.NewMap(sb, types.PubKeysPrefix, "pubkeys", collections.PairKeyCodec(collections.BytesKey, collections.Uint32Key), collections.BytesValue),
		provingSchemes:        collections.NewMap(sb, types.ProvingSchemesPrefix, "proving_schemes", collections.Uint32Key, codec.CollValue[types.ProvingScheme](cdc)),
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

func (k Keeper) SetProvingScheme(ctx context.Context, scheme types.ProvingScheme) error {
	err := k.provingSchemes.Set(ctx, scheme.Index, scheme)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) GetProvingScheme(ctx context.Context, index utils.SEDAKeyIndex) (types.ProvingScheme, error) {
	return k.provingSchemes.Get(ctx, uint32(index))
}

// StartProvingSchemeActivation starts the activation of the given
// proving scheme.
func (k Keeper) StartProvingSchemeActivation(ctx sdk.Context, index utils.SEDAKeyIndex) error {
	scheme, err := k.provingSchemes.Get(ctx, uint32(index))
	if err != nil {
		return err
	}
	scheme.ActivationHeight = ctx.BlockHeight() + ActivationLag
	return k.SetProvingScheme(ctx, scheme)
}

func (k Keeper) IsProvingSchemeActivated(ctx context.Context, index utils.SEDAKeyIndex) (bool, error) {
	scheme, err := k.provingSchemes.Get(ctx, uint32(index))
	if err != nil {
		return false, err
	}
	return scheme.IsActivated, nil
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

		scheme, err := k.provingSchemes.Get(ctx, kv.Key)
		if err != nil {
			return nil, err
		}
		schemes = append(schemes, scheme)
	}
	return schemes, nil
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
