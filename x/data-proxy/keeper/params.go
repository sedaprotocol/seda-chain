package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	return k.params.Set(ctx, params)
}

func (k Keeper) GetMinimumUpdateDelay(ctx sdk.Context) (uint32, error) {
	params, err := k.params.Get(ctx)
	return params.MinFeeUpdateDelay, err
}
