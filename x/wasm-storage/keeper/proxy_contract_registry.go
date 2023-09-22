package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

// SetProxyContractRegistry stores Proxy Contract address.
func (k Keeper) SetProxyContractRegistry(ctx sdk.Context, address sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixProxyContractRegistry, address.Bytes())
}

// GetProxyContractRegistry returns Proxy Contract address.
func (k Keeper) GetProxyContractRegistry(ctx sdk.Context) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixProxyContractRegistry)
	return sdk.AccAddress(bz)
}
