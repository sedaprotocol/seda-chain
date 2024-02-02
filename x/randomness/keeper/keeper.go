package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/sedaprotocol/seda-chain/x/randomness/types"
)

var SeedPrefix = collections.NewPrefix(0)
var ValidatorVRFPrefix = collections.NewPrefix(1)

// GetValidatorVRFKeyFull gets the key for the validator VRF object.
func GetValidatorVRFKeyFull(consensusAddr sdk.ConsAddress) []byte {
	return append(ValidatorVRFPrefix, address.MustLengthPrefix(consensusAddr)...)
}

type Keeper struct {
	Schema              collections.Schema
	Seeds               collections.Item[string]
	ValidatorVRFPubKeys collections.Map[string, cryptotypes.PubKey]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	return &Keeper{
		Seeds:               collections.NewItem(sb, SeedPrefix, "seeds", collections.StringValue),
		ValidatorVRFPubKeys: collections.NewMap(sb, ValidatorVRFPrefix, "validator-vrf-pubkeys", collections.StringKey, codec.CollInterfaceValue[cryptotypes.PubKey](cdc)),
	}
}

// GetSeed returns the seed.
func (k Keeper) GetSeed(ctx sdk.Context) (string, error) {
	seed, err := k.Seeds.Get(ctx)
	if err != nil {
		return "", err
	}

	return seed, nil
}

// SetSeed stores the seed.
func (k Keeper) SetSeed(ctx sdk.Context, seed string) error {
	return k.Seeds.Set(ctx, seed)
}

// GetValidatorVRFPubKey retrieves from the store the VRF public key
// corresponding to the given validator consensus address.
func (k Keeper) GetValidatorVRFPubKey(ctx sdk.Context, consensusAddr string) (cryptotypes.PubKey, error) {
	addr, err := sdk.ConsAddressFromBech32(consensusAddr)
	if err != nil {
		return nil, err
	}
	validatorVRFKeyPrefixFull := GetValidatorVRFKeyFull(addr)
	vrfPubKey, err := k.ValidatorVRFPubKeys.Get(ctx, string(validatorVRFKeyPrefixFull))
	if err != nil {
		return nil, err
	}

	return vrfPubKey, nil
}

func (k Keeper) SetValidatorVRFPubKey(goCtx context.Context, consensusAddr string, vrfPubKey cryptotypes.PubKey) error {
	addr, err := sdk.ConsAddressFromBech32(consensusAddr)
	if err != nil {
		return err
	}
	validatorVRFKeyPrefixFull := GetValidatorVRFKeyFull(addr)
	return k.ValidatorVRFPubKeys.Set(goCtx, string(validatorVRFKeyPrefixFull), vrfPubKey)
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
