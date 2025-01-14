package types

import (
	"context"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

type BatchingKeeper interface {
	SetDataResultForBatching(ctx context.Context, result batchingtypes.DataResult) error
}

type DataProxyKeeper interface {
	GetDataProxyConfig(ctx context.Context, pubKey []byte) (result dataproxytypes.ProxyConfig, err error)
}
