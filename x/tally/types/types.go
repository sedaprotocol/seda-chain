package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

const (
	// ModuleName defines the module name.
	ModuleName = "tally"
)

type WasmStorageKeeper interface {
	GetCoreContractAddr(ctx context.Context) (sdk.AccAddress, error)
	GetOracleProgram(ctx context.Context, hash string) (types.OracleProgram, error)
}
