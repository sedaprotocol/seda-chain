package keeper

import (
	"context"

	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func (q Querier) VRFKeys(ctx context.Context, request *types.VRFKeysRequest) (*types.VRFKeysResponse, error) {
	//TODO implement me
	panic("implement me")
}
