package keeper

import (
	"context"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func (q Querier) DataProxyConfig(ctx context.Context, req *types.QueryDataProxyConfigRequest) (*types.QueryDataProxyConfigResponse, error) {
	result, err := q.GetDataProxyConfig(ctx, req.PubKey)
	if err != nil {
		return nil, err
	}
	return &types.QueryDataProxyConfigResponse{Config: &result}, nil
}
