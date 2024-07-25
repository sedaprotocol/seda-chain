package keeper

import (
	"fmt"

	"cosmossdk.io/log"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

type Keeper struct {
	wasmStorageKeeper types.WasmStorageKeeper
	wasmKeeper        wasmtypes.ContractOpsKeeper
	wasmViewKeeper    wasmtypes.ViewKeeper
}

func NewKeeper(wsk types.WasmStorageKeeper, wk wasmtypes.ContractOpsKeeper, wvk wasmtypes.ViewKeeper) Keeper {
	k := Keeper{
		wasmStorageKeeper: wsk,
		wasmKeeper:        wk,
		wasmViewKeeper:    wvk,
	}
	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
