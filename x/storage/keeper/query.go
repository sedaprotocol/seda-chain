package keeper

import (
	"github.com/sedaprotocol/seda-chain/x/storage/types"
)

var _ types.QueryServer = Keeper{}
