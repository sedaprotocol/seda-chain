package types

import (
	"context"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
)

type BatchingKeeper interface {
	SetDataResultForBatching(ctx context.Context, result batchingtypes.DataResult) error
}
