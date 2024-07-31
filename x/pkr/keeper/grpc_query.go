package keeper

import (
	"context"

	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func (q Querier) ValidatorKeys(ctx context.Context, req *types.QueryValidatorKeysRequest) (*types.QueryValidatorKeysResponse, error) {
	result, err := q.GetValidatorKeys(ctx, req.ValidatorAddr)
	if err != nil {
		return nil, err
	}
	return &types.QueryValidatorKeysResponse{ValidatorPubKeys: result}, nil
}
