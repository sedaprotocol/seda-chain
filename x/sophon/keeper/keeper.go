package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

type Keeper struct {
	bankKeeper types.BankKeeper

	// authority is the address capable of executing MsgUpdateParams. Typically, this should be the gov module address.
	authority string

	Schema collections.Schema
	params collections.Item[types.Params]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, bk types.BankKeeper, authority string) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		authority:  authority,
		bankKeeper: bk,
		params:     collections.NewItem(sb, types.ParamsPrefix, "params", codec.CollValue[types.Params](cdc)),
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

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
