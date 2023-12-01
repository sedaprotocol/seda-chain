package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/sedaprotocol/seda-chain/x/randomness/types"
)

func SeedQueryPlugin(randomnessKeeper *Querier) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {

		var contractQuery types.QuerySeedRequest
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, sdkerrors.Wrap(err, "seed query")
		}

		seedQueryResponse, err := randomnessKeeper.Seed(ctx, &contractQuery)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "seed query")
		}

		bz, err := json.Marshal(seedQueryResponse)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "seed query response")
		}
		return bz, nil

	}
}
