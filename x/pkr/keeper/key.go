package keeper

import (
	"cosmossdk.io/api/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetAllKeys(ctx sdk.Context) ([]crypto.PublicKey, error) {
	return nil, nil
}
