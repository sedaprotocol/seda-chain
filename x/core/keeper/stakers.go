package keeper

import (
	"bytes"
	"encoding/hex"

	"golang.org/x/crypto/sha3"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// GetStaker retrieves a staker given its public key.
func (k Keeper) GetStaker(ctx sdk.Context, pubKey string) (types.Staker, error) {
	staker, err := k.stakers.Get(ctx, pubKey)
	if err != nil {
		return types.Staker{}, err
	}
	return staker, nil
}

// GetExecutors retrieves a list of stakers in the order of their index.
// starting at the offset and
func (k Keeper) GetExecutors(ctx sdk.Context, offset, limit uint32) ([]types.Staker, error) {
	return k.stakers.GetExecutors(ctx, offset, limit)
}

// SetStaker sets a staker in the store.
func (k Keeper) SetStaker(ctx sdk.Context, staker types.Staker) error {
	return k.stakers.Set(ctx, staker.PublicKey, staker)
}

// RemoveStaker removes a staker from the store.
func (k Keeper) RemoveStaker(ctx sdk.Context, pubKey string) error {
	return k.stakers.Remove(ctx, pubKey)
}

// GetStakerCount returns the number of stakers in the store.
func (k Keeper) GetStakerCount(ctx sdk.Context) (uint32, error) {
	return k.stakers.GetStakerCount(ctx)
}

// GetStakerIndex returns the index of a staker given its public key.
func (k Keeper) GetStakerIndex(ctx sdk.Context, pubKey string) (uint32, error) {
	return k.stakers.GetStakerIndex(ctx, pubKey)
}

// GetStakerKey returns the public key of a staker given its index.
func (k Keeper) GetStakerKey(ctx sdk.Context, index uint32) (string, error) {
	return k.stakers.GetStakerKey(ctx, index)
}

func (k Keeper) GetAllStakers(ctx sdk.Context) ([]types.Staker, error) {
	var stakers []types.Staker
	err := k.IterateStakers(ctx, func(staker types.Staker) bool {
		stakers = append(stakers, staker)
		return false
	})
	if err != nil {
		return nil, err
	}
	return stakers, nil
}

func (k Keeper) IterateStakers(ctx sdk.Context, callback func(types.Staker) (stop bool)) error {
	iter, err := k.stakers.Iterate(ctx, nil)
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

func (k Keeper) IsStakerExecutor(ctx sdk.Context, pubKey string) (bool, error) {
	staker, err := k.GetStaker(ctx, pubKey)
	if err != nil {
		return false, err
	}

	stakingConfig, err := k.GetStakingConfig(ctx)
	if err != nil {
		return false, err
	}

	// check if staker is on the allowlist if it's enabled
	if stakingConfig.AllowlistEnabled {
		isAllowlisted, err := k.IsAllowlisted(ctx, pubKey)
		if err != nil {
			return false, err
		}
		if !isAllowlisted {
			return false, nil
		}
	}

	return staker.Staked.GTE(stakingConfig.MinimumStake), nil
}

func computeSelectionHash(pubKey []byte, drID []byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(pubKey)
	hasher.Write(drID)
	return hasher.Sum(nil)
}

func (k Keeper) IsEligibleForDataRequest(ctx sdk.Context, pubKeyBytes, drIDBytes []byte, mininumStake math.Int) (bool, error) {
	dr, err := k.GetDataRequest(ctx, hex.EncodeToString(drIDBytes))
	if err != nil {
		return false, err
	}

	stakers, err := k.stakers.Iterate(ctx, nil)
	if err != nil {
		return false, err
	}
	defer stakers.Close()

	diff := ctx.BlockHeight() - dr.PostedHeight
	var blocksPassed uint64
	if diff > 0 {
		blocksPassed = uint64(diff)
	} else {
		blocksPassed = 0
	}
	drConfig, err := k.GetDataRequestConfig(ctx)
	if err != nil {
		return false, err
	}

	targetHash := computeSelectionHash(pubKeyBytes, drIDBytes)

	totalStakers := 0
	lowerHashCount := uint64(0)
	for ; stakers.Valid(); stakers.Next() {
		staker, err := stakers.Value()
		if err != nil {
			return false, err
		}

		if staker.Staked.LT(mininumStake) {
			continue
		}

		stakerPubKeyBytes, err := hex.DecodeString(staker.PublicKey)
		if err != nil {
			return false, err
		}
		stakerHash := computeSelectionHash(stakerPubKeyBytes, drIDBytes)
		totalStakers++
		if bytes.Compare(stakerHash, targetHash) < 0 {
			lowerHashCount++
		}
	}

	if totalStakers == 0 {
		return false, nil
	}

	var totalNeeded uint64
	backupDelayInBlocks := uint64(drConfig.BackupDelayInBlocks)
	replicationFactor := uint64(dr.ReplicationFactor)
	if blocksPassed > backupDelayInBlocks {
		totalNeeded = replicationFactor + ((blocksPassed - 1) / backupDelayInBlocks)
	} else {
		totalNeeded = replicationFactor
	}

	return lowerHashCount < totalNeeded, nil
}
