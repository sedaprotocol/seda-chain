package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/storage/types"
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

// SetData stores data using its hash as the key.
func (k Keeper) SetData(ctx sdk.Context, data, hash []byte) {
	ctx.KVStore(k.storeKey).Set(types.DataStoreKey(hash), data)
}

// GetData returns data given its key.
func (k Keeper) GetData(ctx sdk.Context, hash []byte) []byte {
	return ctx.KVStore(k.storeKey).Get(types.DataStoreKey(hash))
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
