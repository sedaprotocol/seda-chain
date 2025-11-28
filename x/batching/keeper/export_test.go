/*
   This file is added for test use only.
*/

package keeper

import (
	"context"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/sha3"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	sedatypes "github.com/sedaprotocol/seda-chain/types"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

func (k Keeper) LegacySetDataResultForBatching(ctx context.Context, result types.DataResult) error {
	return k.legacyDataResults.Set(ctx, collections.Join3(false, result.DrId, result.DrBlockHeight), result)
}

func (k Keeper) LegacyConstructBatch(ctx sdk.Context) (types.Batch, types.DataResultTreeEntries, []types.ValidatorTreeEntry, error) {
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
	dataEntries, dataRoot, err := k.LegacyConstructDataResultTree(ctx, newBatchNum)
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

func (k Keeper) LegacyConstructDataResultTree(ctx sdk.Context, newBatchNum uint64) (types.DataResultTreeEntries, []byte, error) {
	dataResults, err := k.GetLegacyDataResults(ctx, false)
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

		err = k.LegacyMarkDataResultAsBatched(ctx, res, newBatchNum)
		if err != nil {
			return types.DataResultTreeEntries{}, nil, err
		}
	}

	return types.DataResultTreeEntries{Entries: entries}, utils.RootFromEntries(treeEntries), nil
}

func (k Keeper) LegacyMarkDataResultAsBatched(ctx context.Context, result types.DataResult, batchNum uint64) error {
	err := k.LegacyRemoveDataResult(ctx, false, result.DrId, result.DrBlockHeight)
	if err != nil {
		return err
	}
	err = k.legacySetDataResultAsBatched(ctx, result)
	if err != nil {
		return err
	}
	return k.SetBatchAssignment(ctx, result.DrId, result.DrBlockHeight, batchNum)
}

func (k Keeper) LegacyRemoveDataResult(ctx context.Context, batched bool, dataReqID string, dataReqHeight uint64) error {
	return k.legacyDataResults.Remove(ctx, collections.Join3(batched, dataReqID, dataReqHeight))
}

func (k Keeper) legacySetDataResultAsBatched(ctx context.Context, result types.DataResult) error {
	return k.legacyDataResults.Set(ctx, collections.Join3(true, result.DrId, result.DrBlockHeight), result)
}
