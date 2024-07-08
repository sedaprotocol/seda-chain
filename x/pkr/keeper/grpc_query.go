package keeper

import (
	"context"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"cosmossdk.io/collections"
	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func (q Querier) KeysByApplication(ctx context.Context, request *types.KeysByApplicationRequest) (*types.KeysByApplicationResponse, error) {
	rng := collections.NewPrefixedPairRange[string, string](request.Application)
	it, err := q.PublicKeys.Iterate(ctx, rng)
	if err != nil {
		return nil, fmt.Errorf("could not iterate in KeysByApplication. err: %w", err)
	}

	kvs, err := it.KeyValues()
	if err != nil {
		return nil, err
	}

	pubKeys := make([]*codectypes.Any, 0)
	for _, kv := range kvs {
		pkAny, err := codectypes.NewAnyWithValue(kv.Value)
		if err != nil {
			return nil, fmt.Errorf("KeysByApplication: err: %w", err)
		}
		pubKeys = append(pubKeys, pkAny)
	}
	return &types.KeysByApplicationResponse{Keys: pubKeys}, nil
}
