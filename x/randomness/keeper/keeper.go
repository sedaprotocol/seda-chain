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

var (
	SeedPrefix         = collections.NewPrefix(0)
	ValidatorVRFPrefix = collections.NewPrefix(1)
)

// GetValidatorVRFKeyPrefixFull gets the key for the validator VRF object.
func GetValidatorVRFKeyPrefixFull(consensusAddr sdk.ConsAddress) []byte {
	return append(ValidatorVRFPrefix, address.MustLengthPrefix(consensusAddr)...)
}

type Keeper struct {
	Schema              collections.Schema
	Seed                collections.Item[string]
	ValidatorVRFPubKeys collections.Map[string, cryptotypes.PubKey]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	return &Keeper{
		Seed:                collections.NewItem(sb, SeedPrefix, "seed", collections.StringValue),
		ValidatorVRFPubKeys: collections.NewMap(sb, ValidatorVRFPrefix, "validator-vrf-pubkeys", collections.StringKey, codec.CollInterfaceValue[cryptotypes.PubKey](cdc)),
	}
}

// GetValidatorVRFPubKey retrieves from the store the VRF public key
// corresponding to the given validator consensus address.
func (k Keeper) GetValidatorVRFPubKey(ctx sdk.Context, consensusAddr string) (cryptotypes.PubKey, error) {
	addr, err := sdk.ConsAddressFromBech32(consensusAddr)
	if err != nil {
		return nil, err
	}
	validatorVRFKeyPrefixFull := GetValidatorVRFKeyPrefixFull(addr)
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
	validatorVRFKeyPrefixFull := GetValidatorVRFKeyPrefixFull(addr)
	return k.ValidatorVRFPubKeys.Set(goCtx, string(validatorVRFKeyPrefixFull), vrfPubKey)
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
