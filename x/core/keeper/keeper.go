package keeper

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

type Keeper struct {
	wasmStorageKeeper types.WasmStorageKeeper
	batchingKeeper    types.BatchingKeeper
	stakingKeeper     types.StakingKeeper
	bankKeeper        types.BankKeeper
	wasmKeeper        wasmtypes.ContractOpsKeeper
	wasmViewKeeper    wasmtypes.ViewKeeper
	authority         string

	Schema    collections.Schema
	Allowlist collections.KeySet[string]
	Stakers   collections.Map[string, types.Staker]
	Params    collections.Item[types.Params]

	DataRequests collections.Map[string, types.DataRequest]
	committing   collections.KeySet[DataRequestIndex]
	// revealing
	// tallying

	// timeoutQueue is a map from timeout heights to data request IDs.
	timeoutQueue collections.Map[int64, string]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, wsk types.WasmStorageKeeper, batk types.BatchingKeeper, sk types.StakingKeeper, bank types.BankKeeper, wk wasmtypes.ContractOpsKeeper, wvk wasmtypes.ViewKeeper, authority string) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		wasmStorageKeeper: wsk,
		batchingKeeper:    batk,
		stakingKeeper:     sk,
		bankKeeper:        bank,
		wasmKeeper:        wk,
		wasmViewKeeper:    wvk,
		authority:         authority,
		Allowlist:         collections.NewKeySet(sb, types.AllowlistKey, "allowlist", collections.StringKey),
		Stakers:           collections.NewMap(sb, types.StakersKey, "stakers", collections.StringKey, codec.CollValue[types.Staker](cdc)),
		Params:            collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		DataRequests:      collections.NewMap(sb, types.DataRequestsKey, "data_requests", collections.StringKey, codec.CollValue[types.DataRequest](cdc)),
		committing:        collections.NewKeySet(sb, types.CommittingKey, "committing", collcodec.NewBytesKey[DataRequestIndex]()),
		timeoutQueue:      collections.NewMap(sb, types.TimeoutQueueKey, "timeout_queue", collections.Int64Key, collections.StringValue),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
