package keeper

import (
	"seda-chain/x/sedachain/types"
)

var _ types.QueryServer = Keeper{}
