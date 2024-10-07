package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

func (k Keeper) setParams(ctx sdk.Context, params types.Params) error {
	return k.params.Set(ctx, params)
}

func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	return k.params.Get(ctx)
}

func (k Keeper) GetValSetTrimPercent(ctx sdk.Context) (uint32, error) {
	params, err := k.params.Get(ctx)
	return params.ValidatorSetTrimPercent, err
}
