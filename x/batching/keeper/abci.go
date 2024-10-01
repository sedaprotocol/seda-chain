package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/cmd/sedad/utils"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
	tallytypes "github.com/sedaprotocol/seda-chain/x/tally/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) (err error) {
	// Use defer to prevent returning an error, which would cause
	// the chain to halt.
	defer func() {
		// Handle a panic.
		if r := recover(); r != nil {
			k.Logger(ctx).Error("recovered from panic in batching EndBlock", "err", r)
		}
		// Handle an error.
		if err != nil {
			k.Logger(ctx).Error("error in batching EndBlock", "err", err)
		}
		err = nil
	}()

	batch, err := k.ConstructBatch(ctx)
	if err != nil {
		return err
	}

	err = k.SetBatch(ctx, batch)
	if err != nil {
		return err
	}
	err = k.IncrementCurrentBatchNum(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) ConstructBatch(ctx sdk.Context) (types.Batch, error) {
	curBatchNum, err := k.GetCurrentBatchNum(ctx)
	if err != nil {
		return types.Batch{}, err
	}
	dataLeaves, dataRootHex, err := k.ConstructDataResultTree(ctx)
	if err != nil {
		return types.Batch{}, err
	}
	valLeaves, valRootHex, err := k.ConstructValidatorTree(ctx)
	if err != nil {
		return types.Batch{}, err
	}

	return types.Batch{
		BatchNumber:      curBatchNum,
		BlockHeight:      ctx.BlockHeight(),
		DataResultRoot:   dataRootHex,
		ValidatorRoot:    valRootHex,
		DataResultLeaves: dataLeaves,
		ValidatorLeaves:  valLeaves,
		BlockTime:        ctx.BlockTime(),
	}, nil
}

// ConstructDataResultTree constructs a data result tree based on the
// batching-ready data results returned from the core contract and
// returns its leaves and its hex-encoded tree root.
func (k Keeper) ConstructDataResultTree(ctx sdk.Context) ([][]byte, string, error) {
	coreContract, err := k.wasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil {
		return nil, "", err
	}
	// TODO: Deal with offset and limits. (#313)
	queryRes, err := k.wasmViewKeeper.QuerySmart(ctx, coreContract, []byte(`{"get_data_results_by_status":{"status": "tallied", "offset": 0, "limit": 100}}`))
	if err != nil {
		return nil, "", err
	}
	if string(queryRes) == "[]" {
		return nil, "", err
	}

	var dataResults []tallytypes.DataResult
	err = json.Unmarshal(queryRes, &dataResults)
	if err != nil {
		return nil, "", err
	}

	leaves := make([][]byte, len(dataResults))
	for _, res := range dataResults {
		resHash, err := hex.DecodeString(res.ID)
		if err != nil {
			return nil, "", err
		}
		leaves = append(leaves, resHash)
	}

	curRoot := utils.RootFromEntries(leaves)
	prevRoot, err := k.GetPreviousDataResultRoot(ctx)
	if err != nil {
		return nil, "", err
	}
	root := utils.RootFromLeaves([][]byte{prevRoot, curRoot})

	// TODO update data result status on contract

	return leaves, hex.EncodeToString(root), nil
}

type validatorData struct {
	secp256k1PubKey cryptotypes.PubKey
	power           math.Int // Validator power multiplied by 10^8.
}

// ConstructValidatorTree constructs a validator tree based on the
// validators in the active set and their registered public keys.
// It returns the tree's entries (unhashed leaf contents) and hex-encoded
// root.
func (k Keeper) ConstructValidatorTree(ctx sdk.Context) ([][]byte, string, error) {
	var validators []validatorData
	totalPower := math.ZeroInt()
	err := k.stakingKeeper.IterateLastValidatorPowers(ctx, func(valAddr sdk.ValAddress, power int64) (stop bool) {
		pubKey, err := k.pubKeyKeeper.GetValidatorKeyAtIndex(ctx, valAddr, utils.SEDAKeysIndexSecp256k1)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				return false
			}
			panic(err)
		}
		validators = append(validators, validatorData{
			secp256k1PubKey: pubKey,
			power:           math.NewInt(power).MulRaw(1e8),
		})
		totalPower = totalPower.AddRaw(power)
		return false
	})
	if err != nil {
		return nil, "", err
	}

	entries := make([][]byte, len(validators))
	for i := range validators {
		// Compute validator voting power percentage.
		powerPercent := validators[i].power.Quo(totalPower).Uint64()

		// An entry is (domain_separator || pubkey || voting_power_percentage).
		separator := []byte("SECP256K1")
		pkBytes := validators[i].secp256k1PubKey.Bytes()

		entry := make([]byte, len(separator)+len(pkBytes)+4)
		copy(entry[:len(separator)], separator)
		copy(entry[len(separator):len(separator)+len(pkBytes)], pkBytes)
		binary.BigEndian.PutUint32(entry[len(separator)+len(pkBytes):], uint32(powerPercent))

		entries[i] = entry
	}

	secp256k1Root := utils.RootFromEntries(entries)
	return entries, hex.EncodeToString(secp256k1Root), nil
}
