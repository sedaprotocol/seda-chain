package keeper

import (
	"github.com/sedaprotocol/seda-chain/x/sedachain/types"
)

var _ types.QueryServer = Keeper{}
