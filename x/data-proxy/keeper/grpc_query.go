package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func (q Querier) DataProxyConfig(ctx context.Context, req *types.QueryDataProxyConfigRequest) (*types.QueryDataProxyConfigResponse, error) {
	result, err := q.GetDataProxyConfig(ctx, req.PubKey)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, types.ErrUnknownDataProxy.Wrapf("no data proxy registered for %s", req.PubKey)
		}

		return nil, err
	}

	return &types.QueryDataProxyConfigResponse{Config: &result}, nil
}
