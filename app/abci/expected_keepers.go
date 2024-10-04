package abci

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

type BatchingKeeper interface {
	GetCurrentBatch(ctx context.Context) (types.Batch, error)
}

type PubKeyKeeper interface {
	GetValidatorKeyAtIndex(ctx context.Context, consensusAddr cryptotypes.Address, index utils.SEDAKeyIndex) (cryptotypes.PubKey, error)
}
