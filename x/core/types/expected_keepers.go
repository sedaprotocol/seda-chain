package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

type BatchingKeeper interface {
	SetDataResultForBatching(ctx context.Context, result batchingtypes.DataResult) error
}

type DataProxyKeeper interface {
	GetDataProxyConfig(ctx context.Context, pubKey []byte) (result dataproxytypes.ProxyConfig, err error)
}

type WasmStorageKeeper interface {
	GetCoreContractAddr(ctx context.Context) (sdk.AccAddress, error)
	GetOracleProgram(ctx context.Context, hash string) (wasmstoragetypes.OracleProgram, error)
}

type StakingKeeper interface {
	BondDenom(ctx context.Context) (string, error)
}

type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
