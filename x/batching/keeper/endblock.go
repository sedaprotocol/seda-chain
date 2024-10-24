package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"errors"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"golang.org/x/crypto/sha3"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
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
		if errors.Is(err, types.ErrNoBatchingUpdate) {
			k.Logger(ctx).Info("skip batch creation", "height", ctx.BlockHeight())
			return nil
		}
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
	var newBatchNum uint64
	var latestDataRoot, latestValRoot string
	latestBatch, err := k.GetLatestBatch(ctx)
	if err != nil {
		if !errors.Is(err, types.ErrBatchingHasNotStarted) {
			return types.Batch{}, nil, nil, err
		}
		newBatchNum = collections.DefaultSequenceStart + 1
	} else {
		newBatchNum = latestBatch.BatchNumber + 1
		latestDataRoot = latestBatch.DataResultRoot
		latestValRoot = latestBatch.ValidatorRoot
	}

	// Compute current data result tree root and the "super root"
	// of current and previous data result trees' roots.
	dataEntries, dataRoot, err := k.ConstructDataResultTree(ctx, newBatchNum)
	if err != nil {
		return types.Batch{}, nil, nil, err
	}
	latestDataRootHex, err := hex.DecodeString(latestDataRoot)
	if err != nil {
		return types.Batch{}, nil, nil, err
	}
	superRoot := utils.RootFromLeaves([][]byte{latestDataRootHex, dataRoot})

	// Compute validator tree root.
	valEntries, valRoot, err := k.ConstructValidatorTree(ctx)
	if err != nil {
		return types.Batch{}, nil, nil, err
	}
	valRootHex := hex.EncodeToString(valRoot)

	// Skip batching if there is no update in data result root nor
	// validator root.
	if len(dataEntries) == 0 && valRootHex == latestValRoot {
		return types.Batch{}, nil, nil, types.ErrNoBatchingUpdate
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
		BatchNumber:           newBatchNum,
		BlockHeight:           ctx.BlockHeight(),
		CurrentDataResultRoot: hex.EncodeToString(dataRoot),
		DataResultRoot:        hex.EncodeToString(superRoot),
		ValidatorRoot:         valRootHex,
		BatchId:               batchID,
		ProvingMedatada:       nil,
	}, dataEntries, valEntries, nil
}

// ConstructDataResultTree constructs a data result tree based on the
// data results that have not been batched yet. It returns the tree's
// entries without the domain separators and the tree root.
func (k Keeper) ConstructDataResultTree(ctx sdk.Context, newBatchNum uint64) ([][]byte, []byte, error) {
	dataResults, err := k.GetDataResults(ctx, false)
	if err != nil {
		return nil, nil, err
	}

	entries := make([][]byte, len(dataResults))
	treeEntries := make([][]byte, len(dataResults))
	for i, res := range dataResults {
		resHash, err := hex.DecodeString(res.Id)
		if err != nil {
			return nil, nil, err
		}
		entries[i] = resHash
		treeEntries[i] = append([]byte{utils.SEDASeparatorDataRequest}, resHash...)

		err = k.markDataResultAsBatched(ctx, res, newBatchNum)
		if err != nil {
			return nil, nil, err
		}
	}

	return entries, utils.RootFromEntries(treeEntries), nil
}

// ConstructValidatorTree constructs a validator tree based on the
// validators in the active set and their registered public keys.
// It returns the tree's entries without the domain separators and
// the tree root.
func (k Keeper) ConstructValidatorTree(ctx sdk.Context) ([][]byte, []byte, error) {
	totalPower, err := k.stakingKeeper.GetLastTotalPower(ctx)
	if err != nil {
		return nil, nil, err
	}

	var entries [][]byte
	var treeEntries [][]byte
	err = k.stakingKeeper.IterateLastValidatorPowers(ctx, func(valAddr sdk.ValAddress, power int64) (stop bool) {
		// Retrieve corresponding public key and convert it to
		// uncompressed form.
		secp256k1PubKey, err := k.pubKeyKeeper.GetValidatorKeyAtIndex(ctx, valAddr, utils.SEDAKeyIndexSecp256k1)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				return false
			}
			panic(err)
		}
		pkBytes, err := decompressPubKey(secp256k1PubKey.Bytes())
		if err != nil {
			k.Logger(ctx).Error("failed to decompress public key", "pubkey", secp256k1PubKey)
			panic(err)
		}

		separator := []byte{utils.SEDASeparatorSecp256k1}
		powerPercent := math.NewInt(power).MulRaw(1e8).Quo(totalPower).Uint64()

		// TODO Validator set trimming

		// An entry is (domain_separator || pubkey || voting_power_percentage).
		treeEntry := make([]byte, len(separator)+len(pkBytes)+4)
		copy(treeEntry[:len(separator)], separator)
		copy(treeEntry[len(separator):len(separator)+len(pkBytes)], pkBytes)
		//nolint:gosec // G115: Max of powerPercent should be 1e8 < 2^64.
		binary.BigEndian.PutUint32(treeEntry[len(separator)+len(pkBytes):], uint32(powerPercent))

		entries = append(entries, treeEntry[len(separator):])
		treeEntries = append(treeEntries, treeEntry)
		return false
	})
	if err != nil {
		return nil, nil, err
	}

	return entries, utils.RootFromEntries(treeEntries), nil
}

// decompressPubKey decompresses a public key in 33-byte compressed
// format into the one in 64-byte uncompressed format.
func decompressPubKey(pubKey []byte) ([]byte, error) {
	x, y := secp256k1.DecompressPubkey(pubKey)
	if x == nil || y == nil {
		return nil, types.ErrInvalidPublicKey
	}
	// 64-byte format: x-coordinate | y-coordinate
	return append(x.Bytes(), y.Bytes()...), nil
}
