package keeper

import (
	"context"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/storage/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func (q Querier) Data(c context.Context, req *types.QueryDataRequest) (*types.QueryDataResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	hash, err := hex.DecodeString(req.Hash)
	if err != nil {
		return nil, err
	}

	data := q.GetData(ctx, hash)

	return &types.QueryDataResponse{
		Data: data,
	}, nil
}
