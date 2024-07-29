package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func (q Querier) ValidatorKeys(ctx context.Context, request *types.QueryValidatorKeysRequest) (*types.QueryValidatorKeysResponse, error) {
	valAddr, err := q.validatorAddressCodec.StringToBytes(request.ValidatorAddr)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}
	rng := collections.NewPrefixedPairRange[[]byte, uint32](valAddr)
	it, err := q.PubKeys.Iterate(ctx, rng)
	if err != nil {
		return nil, fmt.Errorf("could not iterate in ValidatorKeys: %w", err)
	}

	kvs, err := it.KeyValues()
	if err != nil {
		return nil, err
	}
	var results []string
	for _, kv := range kvs {
		results = append(results, fmt.Sprintf("%d,%s", kv.Key.K2(), kv.Value.String()))
	}
	return &types.QueryValidatorKeysResponse{IndexPubkeyPairs: results}, nil
}
