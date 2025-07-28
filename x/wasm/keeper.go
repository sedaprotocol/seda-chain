package wasm

import (
	"context"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type WasmStorageKeeper interface {
	GetCoreContractAddr(ctx context.Context) (sdk.AccAddress, error)
}

type Keeper struct {
	*keeper.Keeper
	WasmStorageKeeper WasmStorageKeeper
}

func NewKeeper(k *keeper.Keeper, wsk WasmStorageKeeper) *Keeper {
	return &Keeper{
		Keeper:            k,
		WasmStorageKeeper: wsk,
	}
}
