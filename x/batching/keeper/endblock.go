package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/sha3"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	sedatypes "github.com/sedaprotocol/seda-chain/types"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) error {
	// Since we're only using the secp256k1 key for batching, we only
	// need to check if the secp256k1 proving scheme is activated.
	isActivated, err := k.pubKeyKeeper.IsProvingSchemeActivated(ctx, sedatypes.SEDAKeyIndexSecp256k1)
	if err != nil {
		return err
	}
	if !isActivated {
		k.Logger(ctx).Info("skip batching since proving scheme has not been activated", "index", sedatypes.SEDAKeyIndexSecp256k1)
		return nil
	}

	batch, dataEntries, valEntries, err := k.ConstructBatch(ctx)
	if err != nil {
		if errors.Is(err, types.ErrNoBatchingUpdate) {
			k.Logger(ctx).Info("skip batch creation due to no update", "height", ctx.BlockHeight())
			return nil
		}
		return err
	}

	err = k.SetNewBatch(ctx, batch, dataEntries, valEntries)
	if err != nil {
		return err
	}

	return k.PruneBatches(ctx)
}

// PruneBatches prunes batches and their associated data based on module
// parameters NumBatchesToKeep and MaxBatchPrunePerBlock.
func (k Keeper) PruneBatches(ctx sdk.Context) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	// Note the current batch number is the number for the next batch.
	currentBatchNum, err := k.GetCurrentBatchNum(ctx)
	if err != nil {
		return err
	}
	if currentBatchNum <= params.NumBatchesToKeep {
		k.Logger(ctx).Info("skip batch pruning", "current_batch_num", currentBatchNum, "num_batches_to_keep", params.NumBatchesToKeep)
		return nil
	}

	iter, err := k.batches.Indexes.Number.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	if !iter.Valid() {
		k.Logger(ctx).Info("nothing was pruned due to invalid iterator - should not happen")
		return nil
	}

	firstKey, err := iter.FullKey()
	if err != nil {
		return err
	}
	firstBatchNumber, firstBatchHeight := firstKey.K1(), firstKey.K2()
	lastBatchNumber := min(firstBatchNumber+params.MaxBatchPrunePerBlock, currentBatchNum-params.NumBatchesToKeep)
	lastBatchHeight, err := k.batchIndex.Get(ctx, lastBatchNumber)
	if err != nil {
		return err
	}

	batchNumRng := new(collections.Range[uint64]).StartInclusive(firstBatchNumber).EndExclusive(lastBatchNumber)
	err = k.batchIndex.Clear(ctx, batchNumRng)
	if err != nil {
		return err
	}
	err = k.dataResultTreeEntries.Clear(ctx, batchNumRng)
	if err != nil {
		return err
	}

	batchHeightRng := new(collections.Range[int64]).StartInclusive(firstBatchHeight).EndExclusive(lastBatchHeight)
	err = k.batchesMap.Clear(ctx, batchHeightRng)
	if err != nil {
		return err
	}

	valRng := new(collections.Range[collections.Pair[uint64, []byte]]).
		StartInclusive(collections.PairPrefix[uint64, []byte](firstBatchNumber)).
		EndInclusive(collections.PairPrefix[uint64, []byte](lastBatchNumber))
	err = k.validatorTreeEntries.Clear(ctx, valRng)
	if err != nil {
		return err
	}
	err = k.batchSignatures.Clear(ctx, valRng)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info("successfully pruned batch data")
	return nil
}

// ConstructBatch constructs a data result tree from unbatched data
// results and a validator tree from the current active validator set.
// It returns a resulting batch, data result tree entries, and validator
// tree entries in that order.
func (k Keeper) ConstructBatch(ctx sdk.Context) (types.Batch, types.DataResultTreeEntries, []types.ValidatorTreeEntry, error) {
	var newBatchNum uint64
	var latestDataRootHex, latestValRootHex string
	latestBatch, err := k.GetLatestBatch(ctx)
	if err != nil {
		if !errors.Is(err, types.ErrBatchingHasNotStarted) {
			return types.Batch{}, types.DataResultTreeEntries{}, nil, err
		}
		newBatchNum = collections.DefaultSequenceStart
	} else {
		newBatchNum = latestBatch.BatchNumber + 1
		latestDataRootHex = latestBatch.DataResultRoot
		latestValRootHex = latestBatch.ValidatorRoot
	}

	// Compute current data result tree root and the "super root"
	// of current and previous data result trees' roots.
	dataEntries, dataRoot, err := k.ConstructDataResultTree(ctx, newBatchNum)
	if err != nil {
		return types.Batch{}, types.DataResultTreeEntries{}, nil, err
	}
	latestDataRoot, err := hex.DecodeString(latestDataRootHex)
	if err != nil {
		return types.Batch{}, types.DataResultTreeEntries{}, nil, err
	}
	superRoot := utils.RootFromLeaves([][]byte{latestDataRoot, dataRoot})

	// Compute validator tree root.
	valEntries, valRoot, err := k.ConstructValidatorTree(ctx)
	if err != nil {
		return types.Batch{}, types.DataResultTreeEntries{}, nil, err
	}
	valRootHex := hex.EncodeToString(valRoot)

	// Skip batching if there is no update in data result root nor
	// validator root.
	if len(dataEntries.Entries) == 0 && valRootHex == latestValRootHex {
		return types.Batch{}, types.DataResultTreeEntries{}, nil, types.ErrNoBatchingUpdate
	}

	var provingMetaData, provingMetaDataHash []byte
	if len(provingMetaData) == 0 {
		provingMetaDataHash = make([]byte, 32) // zero hash
	} else {
		hasher := sha3.NewLegacyKeccak256()
		hasher.Write(provingMetaData)
		provingMetaDataHash = hasher.Sum(nil)
	}

	batchID := types.ComputeBatchID(newBatchNum, ctx.BlockHeight(), valRoot, superRoot, provingMetaDataHash)

	return types.Batch{
		BatchNumber:           newBatchNum,
		BlockHeight:           ctx.BlockHeight(),
		CurrentDataResultRoot: hex.EncodeToString(dataRoot),
		DataResultRoot:        hex.EncodeToString(superRoot),
		ValidatorRoot:         valRootHex,
		BatchId:               batchID,
		ProvingMetadata:       provingMetaData,
	}, dataEntries, valEntries, nil
}

// ConstructDataResultTree constructs a data result tree based on the
// data results that have not been batched yet. It returns the tree's
// entries without the domain separators and the tree root.
func (k Keeper) ConstructDataResultTree(ctx sdk.Context, newBatchNum uint64) (types.DataResultTreeEntries, []byte, error) {
	dataResults, err := k.GetDataResults(ctx, false)
	if err != nil {
		return types.DataResultTreeEntries{}, nil, err
	}

	entries := make([][]byte, len(dataResults))
	treeEntries := make([][]byte, len(dataResults))
	for i, res := range dataResults {
		resID, err := hex.DecodeString(res.Id)
		if err != nil {
			return types.DataResultTreeEntries{}, nil, err
		}
		entries[i] = resID
		treeEntries[i] = append([]byte{sedatypes.SEDASeparatorDataResult}, resID...)

		err = k.MarkDataResultAsBatched(ctx, res, newBatchNum)
		if err != nil {
			return types.DataResultTreeEntries{}, nil, err
		}
	}

	return types.DataResultTreeEntries{Entries: entries}, utils.RootFromEntries(treeEntries), nil
}

// ConstructValidatorTree constructs a validator tree based on the
// validators in the active set and their registered public keys.
// It returns the tree's entries without the domain separators, batch
// signature entries with validator address and public key fields
// populated, and the tree root.
func (k Keeper) ConstructValidatorTree(ctx sdk.Context) ([]types.ValidatorTreeEntry, []byte, error) {
	totalPower, err := k.stakingKeeper.GetLastTotalPower(ctx)
	if err != nil {
		return nil, nil, err
	}

	var entries []types.ValidatorTreeEntry
	var treeEntries [][]byte
	err = k.stakingKeeper.IterateLastValidatorPowers(ctx, func(valAddr sdk.ValAddress, power int64) (stop bool) {
		// Retrieve corresponding public key and convert it to
		// uncompressed form.
		secp256k1PubKey, err := k.pubKeyKeeper.GetValidatorKeyAtIndex(ctx, valAddr, sedatypes.SEDAKeyIndexSecp256k1)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				return false
			}
			panic(err)
		}
		ethAddr, err := utils.PubKeyToEthAddress(secp256k1PubKey)
		if err != nil {
			k.Logger(ctx).Error("failed to decompress public key", "pubkey", secp256k1PubKey)
			panic(err)
		}

		separator := []byte{sedatypes.SEDASeparatorSecp256k1}
		//nolint:gosec // G115: Max of powerPercent should be 1e8 < 2^64.
		powerPercent := uint32(math.NewInt(power).MulRaw(1e8).Quo(totalPower).Uint64())

		// A tree entry is (domain_separator | address | voting_power_percentage).
		treeEntry := make([]byte, len(separator)+len(ethAddr)+4)
		copy(treeEntry[:len(separator)], separator)
		copy(treeEntry[len(separator):len(separator)+len(ethAddr)], ethAddr)
		binary.BigEndian.PutUint32(treeEntry[len(separator)+len(ethAddr):], powerPercent)

		treeEntries = append(treeEntries, treeEntry)

		entries = append(entries, types.ValidatorTreeEntry{
			ValidatorAddress:   valAddr.Bytes(),
			VotingPowerPercent: powerPercent,
			EthAddress:         ethAddr,
		})

		return false
	})
	if err != nil {
		return nil, nil, err
	}

	return entries, utils.RootFromEntries(treeEntries), nil
}
