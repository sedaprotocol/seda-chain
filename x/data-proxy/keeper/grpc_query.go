package keeper

import (
	"context"
	"encoding/hex"
	"errors"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func (q Querier) DataProxyConfig(ctx context.Context, req *types.QueryDataProxyConfigRequest) (*types.QueryDataProxyConfigResponse, error) {
	pubKeyBytes, err := hex.DecodeString(req.PubKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", req.PubKey)
	}

	result, err := q.GetDataProxyConfig(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no data proxy registered for %s", req.PubKey)
		}

		return nil, err
	}

	return &types.QueryDataProxyConfigResponse{Config: &result}, nil
}
