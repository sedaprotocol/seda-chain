package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/randomness/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// GetSeed returns the seed.
func (k Keeper) GetSeed(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixSeed)
	return string(bz)
}

// SetSeed stores the seed.
func (k Keeper) SetSeed(ctx sdk.Context, seed string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixSeed, []byte(seed))
}

// GetValidatorVRFPubKey retrieves from the store the VRF public key
// corresponding to the given validator consensus address.
func (k Keeper) GetValidatorVRFPubKey(ctx sdk.Context, consensusAddr string) (cryptotypes.PubKey, error) {
	addr, err := sdk.ConsAddressFromBech32(consensusAddr)
	if err != nil {
		return nil, err
	}

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetValidatorVRFKey(addr))
	if bz == nil {
		return nil, fmt.Errorf("vrf pubkey not found for %s", consensusAddr)
	}

	var vrfPubKey cryptotypes.PubKey
	err = k.cdc.UnmarshalInterface(bz, &vrfPubKey)
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

	store := sdk.UnwrapSDKContext(goCtx).KVStore(k.storeKey)
	bz, err := k.cdc.MarshalInterface(vrfPubKey)
	if err != nil {
		return err
	}
	store.Set(types.GetValidatorVRFKey(addr), bz)
	return nil
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
