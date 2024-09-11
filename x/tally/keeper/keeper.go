package keeper

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

type Keeper struct {
	wasmStorageKeeper types.WasmStorageKeeper
	batchingKeeper    types.BatchingKeeper
	wasmKeeper        wasmtypes.ContractOpsKeeper
	wasmViewKeeper    wasmtypes.ViewKeeper
}

func NewKeeper(wsk types.WasmStorageKeeper, bk types.BatchingKeeper, wk wasmtypes.ContractOpsKeeper, wvk wasmtypes.ViewKeeper) Keeper {
	k := Keeper{
		wasmStorageKeeper: wsk,
		batchingKeeper:    bk,
		wasmKeeper:        wk,
		wasmViewKeeper:    wvk,
	}
	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
