package wasm

import (
	"context"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"

	corestoretypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//nolint:revive
type WasmStorageKeeper interface {
	GetCoreContractAddr(ctx context.Context) (sdk.AccAddress, error)
}

type StakingKeeper interface {
	BondDenom(ctx context.Context) (string, error)
}

type Keeper struct {
	*keeper.Keeper                                  // default wasm keeper
	WasmStorageKeeper WasmStorageKeeper             // for core contract shim
	StakingKeeper     StakingKeeper                 // for core contract shim
	cdc               codec.Codec                   // for core contract shim
	router            keeper.MessageRouter          // for core contract shim
	storeService      corestoretypes.KVStoreService // for core contract query shim
	queryGasLimit     uint64                        // for core contract query shim
	queryRouter       keeper.GRPCQueryRouter        // for core contract query shim
}

func NewKeeper(
	k *keeper.Keeper,
	sk StakingKeeper,
	cdc codec.Codec,
	router keeper.MessageRouter,
	queryRouter keeper.GRPCQueryRouter,
	storeService corestoretypes.KVStoreService,
) *Keeper {
	return &Keeper{
		Keeper:        k,
		StakingKeeper: sk,
		cdc:           cdc,
		router:        router,
		storeService:  storeService,
		queryGasLimit: k.QueryGasLimit(),
		queryRouter:   queryRouter,
	}
}

func (k *Keeper) SetWasmStorageKeeper(wsk WasmStorageKeeper) {
	k.WasmStorageKeeper = wsk
}

func (k *Keeper) SetRouter(router keeper.MessageRouter) {
	k.router = router
}
