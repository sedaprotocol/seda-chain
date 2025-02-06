package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

type Keeper struct {
	accountKeeper  types.AccountKeeper
	bankKeeper     types.BankKeeper
	wasmKeeper     wasmtypes.ContractOpsKeeper
	wasmViewKeeper wasmtypes.ViewKeeper

	// authority is the address capable of executing MsgUpdateParams
	// or MsgStoreExecutorWasm. Typically, this should be the gov module
	// address.
	authority string

	Schema                  collections.Schema
	OracleProgram           collections.Map[[]byte, types.OracleProgram]
	ExecutorWasm            collections.Map[[]byte, types.ExecutorWasm]
	OracleProgramExpiration collections.KeySet[collections.Pair[int64, []byte]]
	CoreContractRegistry    collections.Item[string]
	Params                  collections.Item[types.Params]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	authority string,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	wk wasmtypes.ContractOpsKeeper,
	wvk wasmtypes.ViewKeeper,
) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		authority:               authority,
		accountKeeper:           ak,
		bankKeeper:              bk,
		wasmKeeper:              wk,
		wasmViewKeeper:          wvk,
		OracleProgram:           collections.NewMap(sb, types.OracleProgramPrefix, "oracle_program", collections.BytesKey, codec.CollValue[types.OracleProgram](cdc)),
		ExecutorWasm:            collections.NewMap(sb, types.ExecutorPrefix, "executor_wasm", collections.BytesKey, codec.CollValue[types.ExecutorWasm](cdc)),
		OracleProgramExpiration: collections.NewKeySet(sb, types.WasmExpPrefix, "oracle_program_expiration", collections.PairKeyCodec(collections.Int64Key, collections.BytesKey)),
		CoreContractRegistry:    collections.NewItem(sb, types.CoreContractRegistryPrefix, "core_contract_registry", collections.StringValue),
		Params:                  collections.NewItem(sb, types.ParamsPrefix, "params", codec.CollValue[types.Params](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return &k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetCoreContractAddr retrieves the core contract address.
func (k Keeper) GetCoreContractAddr(ctx context.Context) (sdk.AccAddress, error) {
	contractAddrBech32, err := k.CoreContractRegistry.Get(ctx)
	if err != nil {
		return nil, err
	}
	if contractAddrBech32 == "" {
		return nil, nil
	}
	contractAddr, err := sdk.AccAddressFromBech32(contractAddrBech32)
	if err != nil {
		return nil, err
	}
	return contractAddr, nil
}

// GetOracleProgram retrieves the oracle program from the store
// given its hex-encoded hash.
func (k Keeper) GetOracleProgram(ctx context.Context, hash string) (types.OracleProgram, error) {
	hexHash, err := hex.DecodeString(hash)
	if err != nil {
		return types.OracleProgram{}, types.ErrInvalidHexWasmHash
	}
	program, err := k.OracleProgram.Get(ctx, hexHash)
	if err != nil {
		return types.OracleProgram{}, err
	}
	return program, nil
}

// IterateOraclePrograms iterates over the oracle programs and
// performs a given callback function.
func (k Keeper) IterateOraclePrograms(ctx sdk.Context, callback func(program types.OracleProgram) (stop bool)) error {
	iter, err := k.OracleProgram.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		if callback(kv.Value) {
			break
		}
	}
	return nil
}

// IterateExecutorWasms iterates over the executor wasms and performs
// a given callback function.
func (k Keeper) IterateExecutorWasms(ctx sdk.Context, callback func(wasm types.ExecutorWasm) (stop bool)) error {
	iter, err := k.ExecutorWasm.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		if callback(kv.Value) {
			break
		}
	}
	return nil
}

// ListOraclePrograms returns hashes and expiration block heights
// of all oracle programs in the store.
func (k Keeper) ListOraclePrograms(ctx sdk.Context) []string {
	var results []string
	err := k.IterateOraclePrograms(ctx, func(w types.OracleProgram) bool {
		results = append(results, fmt.Sprintf("%s,%d", hex.EncodeToString(w.Hash), w.ExpirationHeight))
		return false
	})
	if err != nil {
		return nil
	}
	return results
}

// ListExecutorWasms returns hex-encoded hashes of all executor wasms in the store.
func (k Keeper) ListExecutorWasms(ctx sdk.Context) []string {
	var hashes []string
	err := k.IterateExecutorWasms(ctx, func(w types.ExecutorWasm) bool {
		hashes = append(hashes, hex.EncodeToString(w.Hash))
		return false
	})
	if err != nil {
		return nil
	}
	return hashes
}

func (k Keeper) GetAllOraclePrograms(ctx sdk.Context) []types.OracleProgram {
	var programs []types.OracleProgram
	err := k.IterateOraclePrograms(ctx, func(program types.OracleProgram) bool {
		programs = append(programs, program)
		return false
	})
	if err != nil {
		return nil
	}
	return programs
}

func (k Keeper) GetAllExecutorWasms(ctx sdk.Context) []types.ExecutorWasm {
	var wasms []types.ExecutorWasm
	err := k.IterateExecutorWasms(ctx, func(wasm types.ExecutorWasm) bool {
		wasms = append(wasms, wasm)
		return false
	})
	if err != nil {
		return nil
	}
	return wasms
}

// GetExpiredOracleProgamKeys retrieves the keys of the oracle programs
// that will expire at the given block height.
func (k Keeper) GetExpiredOracleProgamKeys(ctx sdk.Context, expirationHeight int64) ([][]byte, error) {
	ret := make([][]byte, 0)
	rng := collections.NewPrefixedPairRange[int64, []byte](expirationHeight)

	itr, err := k.OracleProgramExpiration.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}

	keys, err := itr.Keys()
	if err != nil {
		return nil, err
	}
	for _, k := range keys {
		ret = append(ret, k.K2())
	}
	return ret, nil
}

func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	return k.Params.Get(ctx)
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
