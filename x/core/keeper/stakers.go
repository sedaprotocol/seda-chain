package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (k Keeper) GetStakersCount(ctx sdk.Context) (int, error) {
	count := 0
	err := k.Stakers.Walk(ctx, nil, func(key string, value types.Staker) (stop bool, err error) {
		count++
		return false, nil
	})
	return count, err
}
