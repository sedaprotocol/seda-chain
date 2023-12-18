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

func (k Keeper) GetValidatorVRFPubKey(ctx sdk.Context, consAddr string) (cryptotypes.PubKey, error) {
	addr, err := sdk.ConsAddressFromBech32(consAddr)
	if err != nil {
		return nil, err
	}

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetValidatorVRFKey(addr))
	if bz == nil {
		return nil, fmt.Errorf("vrf pubkey not found for %s", consAddr)
	}

	var vrfPubKey cryptotypes.PubKey
	err = k.cdc.UnmarshalInterface(bz, &vrfPubKey)
	if err != nil {
		return nil, err
	}

	return vrfPubKey, nil
}

func (k Keeper) SetValidatorVRFPubKey(goCtx context.Context, consAddr string, vrfPubKey cryptotypes.PubKey) error {
	addr, err := sdk.ConsAddressFromBech32(consAddr)
	if err != nil {
		return err
	}
	fmt.Println(addr.Bytes())
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)

	bz, err := k.cdc.MarshalInterface(vrfPubKey)
	if err != nil {
		return err
	}
	store.Set(types.GetValidatorVRFKey(addr.Bytes()), bz)

	return nil
}

// func (k Keeper) GetValidatorVRFPubKey(ctx sdk.Context, addr sdk.ValAddress) (cryptotypes.PubKey, error) {
// 	var valVrf types.ValidatorVRF
// 	store := ctx.KVStore(k.storeKey)
// 	bz := store.Get(types.GetValidatorVRFKey(addr))
// 	if bz == nil {
// 		return nil, fmt.Errorf("VRF object not found for validator %s", addr.String())
// 	}
// 	k.cdc.MustUnmarshal(bz, &valVrf)

// 	pk, ok := valVrf.VrfPubkey.GetCachedValue().(cryptotypes.PubKey)
// 	if !ok {
// 		return nil, fmt.Errorf("expecting cryptotypes.PubKey, got %T", pk)
// 	}
// 	return pk, nil
// }

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
