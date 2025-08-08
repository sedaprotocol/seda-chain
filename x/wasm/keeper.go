package wasm

import (
	"context"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type WasmStorageKeeper interface {
	GetCoreContractAddr(ctx context.Context) (sdk.AccAddress, error)
}

type StakingKeeper interface {
	BondDenom(ctx context.Context) (string, error)
}

type Keeper struct {
	*keeper.Keeper                         // default wasm keeper
	WasmStorageKeeper WasmStorageKeeper    // for core contract shim
	StakingKeeper     StakingKeeper        // for core contract shim
	cdc               codec.Codec          // for core contract shim
	router            keeper.MessageRouter // for core contract shim
}

func NewKeeper(k *keeper.Keeper, sk StakingKeeper, cdc codec.Codec, router keeper.MessageRouter) *Keeper {
	return &Keeper{
		Keeper:        k,
		StakingKeeper: sk,
		cdc:           cdc,
		router:        router,
	}
}

func (k *Keeper) SetWasmStorageKeeper(wsk WasmStorageKeeper) {
	k.WasmStorageKeeper = wsk
}

func (k *Keeper) SetRouter(router keeper.MessageRouter) {
	k.router = router
}
