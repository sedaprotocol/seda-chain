package keeper

import (
	"encoding/binary"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DataRequestIndex is a 56-byte index for data requests used to sort them by
// their posted gas prices and heights.
// 0                  16       24      56 (byte)
// | posted_gas_price | height | dr_id |
type DataRequestIndex []byte

func NewDataRequestIndex(drID string, gasPrice math.Int, height uint64) DataRequestIndex {
	// Treat gasPrice as a 128-bit unsigned integer.
	priceBytes := make([]byte, 16)
	gasPrice.BigInt().FillBytes(priceBytes)

	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, height)

	drIDBytes := []byte(drID)
	return append(append(priceBytes, heightBytes...), drIDBytes...)
}

func (k Keeper) AddToCommitting(ctx sdk.Context, index DataRequestIndex) error {
	return k.committing.Set(ctx, index)
}
