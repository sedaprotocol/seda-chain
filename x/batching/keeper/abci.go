package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/sha3"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/cmd/sedad/utils"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) (err error) {
	// Use defer to prevent returning an error, which would cause
	// the chain to halt.
	defer func() {
		// Handle a panic.
		if r := recover(); r != nil {
			k.Logger(ctx).Error("recovered from panic in batching end blocker", "err", r)
		}
		// Handle an error.
		if err != nil {
			k.Logger(ctx).Error("error in batching end blocker", "err", err)
		}
		err = nil
	}()

	batch, dataEntries, valEntries, err := k.ConstructBatch(ctx)
	if err != nil {
		return err
	}

	err = k.SetNewBatch(ctx, batch, dataEntries, valEntries)
	if err != nil {
		return err
	}
	return nil
}

// ConstructBatch constructs a data result tree from unbatched data
// results and a validator tree from the current active validator set.
// It returns a resulting batch, data result tree entries, and validator
// tree entries in that order.
func (k Keeper) ConstructBatch(ctx sdk.Context) (types.Batch, [][]byte, [][]byte, error) {
	// Note current will be "old" for this new batch.
	oldBatchNum, err := k.GetCurrentBatchNum(ctx)
	if err != nil {
		return types.Batch{}, nil, nil, err
	}
	newBatchNum := oldBatchNum + 1

	dataEntries, dataRoot, err := k.ConstructDataResultTree(ctx)
	if err != nil {
		return types.Batch{}, nil, nil, err
	}
	valEntries, valRoot, err := k.ConstructValidatorTree(ctx)
	if err != nil {
		return types.Batch{}, nil, nil, err
	}

	// Compute the batch ID, which is defined as
	// keccak256(batch_number, block_height, validator_root, results_root).
	var hashContent []byte
	hashContent = binary.BigEndian.AppendUint64(hashContent, newBatchNum)
	//nolint:gosec // G115: We shouldn't get negative block heights anyway.
	hashContent = binary.BigEndian.AppendUint64(hashContent, uint64(ctx.BlockHeight()))
	hashContent = append(hashContent, valRoot...)
	hashContent = append(hashContent, dataRoot...)

	hash := sha3.NewLegacyKeccak256()
	hash.Write(hashContent)
	batchID := hash.Sum(nil)

	return types.Batch{
		BatchNumber:     newBatchNum,
		BlockHeight:     ctx.BlockHeight(),
		DataResultRoot:  hex.EncodeToString(dataRoot),
		ValidatorRoot:   hex.EncodeToString(valRoot),
		BatchId:         batchID,
		ProvingMedatada: nil,
	}, dataEntries, valEntries, nil
}

// ConstructDataResultTree constructs a data result tree based on the
// data results that have not been batched yet. It returns the tree's
// entries (unhashed leaf contents) and hex-encoded root.
func (k Keeper) ConstructDataResultTree(ctx sdk.Context) ([][]byte, []byte, error) {
	dataResults, err := k.GetDataResults(ctx, false)
	if err != nil {
		return nil, nil, err
	}

	entries := make([][]byte, len(dataResults))
	for i, res := range dataResults {
		resHash, err := hex.DecodeString(res.Id)
		if err != nil {
			return nil, nil, err
		}
		entries[i] = resHash

		err = k.markDataResultAsBatched(ctx, res)
		if err != nil {
			return nil, nil, err
		}
	}

	newRoot := utils.RootFromEntries(entries)
	curRoot, err := k.GetLatestDataResultRoot(ctx)
	if err != nil {
		return nil, nil, err
	}
	root := utils.RootFromLeaves([][]byte{curRoot, newRoot})

	return entries, root, nil
}

// ConstructValidatorTree constructs a validator tree based on the
// validators in the active set and their registered public keys.
// It returns the tree's entries (unhashed leaf contents) and hex-encoded
// root.
func (k Keeper) ConstructValidatorTree(ctx sdk.Context) ([][]byte, []byte, error) {
	totalPower, err := k.stakingKeeper.GetLastTotalPower(ctx)
	if err != nil {
		return nil, nil, err
	}

	var entries [][]byte
	err = k.stakingKeeper.IterateLastValidatorPowers(ctx, func(valAddr sdk.ValAddress, power int64) (stop bool) {
		secp256k1PubKey, err := k.pubKeyKeeper.GetValidatorKeyAtIndex(ctx, valAddr, utils.SEDAKeysIndexSecp256k1)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				return false
			}
			panic(err)
		}

		// Compute validator voting power percentage.
		powerPercent := math.NewInt(power).MulRaw(1e8).Quo(totalPower).Uint64()

		// TODO Validator set trimming

		// An entry is (domain_separator || pubkey || voting_power_percentage).
		separator := []byte("SECP256K1")
		pkBytes := secp256k1PubKey.Bytes()

		entry := make([]byte, len(separator)+len(pkBytes)+4)
		copy(entry[:len(separator)], separator)
		copy(entry[len(separator):len(separator)+len(pkBytes)], pkBytes)
		//nolint:gosec // G115: Max of powerPercent should be 1e8 < 2^64.
		binary.BigEndian.PutUint32(entry[len(separator)+len(pkBytes):], uint32(powerPercent))

		entries = append(entries, entry)
		return false
	})
	if err != nil {
		return nil, nil, err
	}

	secp256k1Root := utils.RootFromEntries(entries)
	return entries, secp256k1Root, nil
}
