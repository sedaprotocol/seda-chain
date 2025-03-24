package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sedatypes "github.com/sedaprotocol/seda-chain/types"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

type Keeper struct {
	stakingKeeper         types.StakingKeeper
	slashingKeeper        types.SlashingKeeper
	validatorAddressCodec address.Codec
	authority             string

	Schema         collections.Schema
	pubKeys        collections.Map[collections.Pair[[]byte, uint32], []byte]
	provingSchemes collections.Map[uint32, types.ProvingScheme]
	params         collections.Item[types.Params]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, stk types.StakingKeeper, slk types.SlashingKeeper, valAddrCdc address.Codec, authority string) *Keeper {
	if valAddrCdc == nil {
		panic("validator address codec is nil")
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		stakingKeeper:         stk,
		slashingKeeper:        slk,
		validatorAddressCodec: valAddrCdc,
		pubKeys:               collections.NewMap(sb, types.PubKeysPrefix, "pubkeys", collections.PairKeyCodec(collections.BytesKey, collections.Uint32Key), collections.BytesValue),
		provingSchemes:        collections.NewMap(sb, types.ProvingSchemesPrefix, "proving_schemes", collections.Uint32Key, codec.CollValue[types.ProvingScheme](cdc)),
		params:                collections.NewItem(sb, types.ParamsPrefix, "params", codec.CollValue[types.Params](cdc)),
		authority:             authority,
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

// StoreIndexedPubKeys stores the given list of indexed public keys
// for a validator.
func (k Keeper) StoreIndexedPubKeys(ctx sdk.Context, valAddr sdk.ValAddress, pubKeys []types.IndexedPubKey) error {
	for _, pk := range pubKeys {
		err := k.SetValidatorKeyAtIndex(ctx, valAddr, sedatypes.SEDAKeyIndex(pk.Index), pk.PubKey)
		if err != nil {
			return err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeAddKey,
				sdk.NewAttribute(types.AttributeValidatorAddr, valAddr.String()),
				sdk.NewAttribute(types.AttributePubKeyIndex, fmt.Sprintf("%d", pk.Index)),
				sdk.NewAttribute(types.AttributePublicKey, hex.EncodeToString(pk.PubKey)),
			),
		)
	}
	return nil
}

func (k Keeper) SetValidatorKeyAtIndex(ctx context.Context, validatorAddr sdk.ValAddress, index sedatypes.SEDAKeyIndex, pubKey []byte) error {
	err := k.pubKeys.Set(ctx, collections.Join(validatorAddr.Bytes(), uint32(index)), pubKey)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) GetValidatorKeyAtIndex(ctx context.Context, validatorAddr sdk.ValAddress, index sedatypes.SEDAKeyIndex) ([]byte, error) {
	pubKey, err := k.pubKeys.Get(ctx, collections.Join(validatorAddr.Bytes(), uint32(index)))
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

// HasRegisteredKey returns true if the validator has registered a key
// at the index.
func (k Keeper) HasRegisteredKey(ctx context.Context, validatorAddr sdk.ValAddress, index sedatypes.SEDAKeyIndex) (bool, error) {
	_, err := k.pubKeys.Get(ctx, collections.Join(validatorAddr.Bytes(), uint32(index)))
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
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

func (k Keeper) GetProvingScheme(ctx context.Context, index sedatypes.SEDAKeyIndex) (types.ProvingScheme, error) {
	return k.provingSchemes.Get(ctx, uint32(index))
}

// StartProvingSchemeActivation starts the activation of the given
// proving scheme.
func (k Keeper) StartProvingSchemeActivation(ctx sdk.Context, index sedatypes.SEDAKeyIndex) error {
	scheme, err := k.provingSchemes.Get(ctx, uint32(index))
	if err != nil {
		return err
	}
	activationBlockDelay, err := k.GetActivationBlockDelay(ctx)
	if err != nil {
		return err
	}
	scheme.ActivationHeight = ctx.BlockHeight() + activationBlockDelay
	return k.SetProvingScheme(ctx, scheme)
}

func (k Keeper) CancelProvingSchemeActivation(ctx sdk.Context, index sedatypes.SEDAKeyIndex) error {
	scheme, err := k.provingSchemes.Get(ctx, uint32(index))
	if err != nil {
		return err
	}
	scheme.ActivationHeight = types.DefaultActivationHeight
	return k.SetProvingScheme(ctx, scheme)
}

func (k Keeper) IsProvingSchemeActivated(ctx context.Context, index sedatypes.SEDAKeyIndex) (bool, error) {
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

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	return k.params.Set(ctx, params)
}

func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	return k.params.Get(ctx)
}

func (k Keeper) GetActivationBlockDelay(ctx sdk.Context) (int64, error) {
	params, err := k.params.Get(ctx)
	if err != nil {
		return 0, err
	}
	return params.ActivationBlockDelay, nil
}

func (k Keeper) GetActivationThresholdPercent(ctx sdk.Context) (uint32, error) {
	params, err := k.params.Get(ctx)
	if err != nil {
		return 0, err
	}
	return params.ActivationThresholdPercent, nil
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
