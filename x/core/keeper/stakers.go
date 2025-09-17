package keeper

import (
	"bytes"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/crypto/sha3"

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

// SetStaker sets a staker in the store.
func (k Keeper) SetStaker(ctx sdk.Context, staker types.Staker) error {
	return k.stakers.Set(ctx, staker.PublicKey, staker)
}

// RemoveStaker removes a staker from the store.
func (k Keeper) RemoveStaker(ctx sdk.Context, pubKey string) error {
	return k.stakers.Remove(ctx, pubKey)
}

// GetStakersCount returns the number of stakers in the store.
func (k Keeper) GetStakersCount(ctx sdk.Context) (uint32, error) {
	count := uint32(0)
	err := k.stakers.Walk(ctx, nil, func(_ string, _ types.Staker) (stop bool, err error) {
		count++
		return false, nil
	})
	return count, err
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

func (k Keeper) IsEligibleForDr(ctx sdk.Context, pubKeyBytes, drIdBytes []byte, dr types.DataRequest) (bool, error) {
	stakingConfig, err := k.GetStakingConfig(ctx)
	if err != nil {
		return false, err
	}

	stakers, err := k.stakers.Iterate(ctx, nil)
	if err != nil {
		return false, err
	}
	defer stakers.Close()

	blocksPassed := uint64(ctx.BlockHeight()) - uint64(dr.PostedHeight)
	drConfig, err := k.GetDataRequestConfig(ctx)
	if err != nil {
		return false, err
	}

	targetHash := computeSelectionHash(pubKeyBytes, drIdBytes)

	totalStakers := 0
	lowerHashCount := uint64(0)
	for ; stakers.Valid(); stakers.Next() {
		staker, err := stakers.Value()
		if err != nil {
			return false, err
		}

		if !staker.Staked.LT(stakingConfig.MinimumStake) {
			continue
		}

		// TODO: we should store pubkey as bytes to avoid this
		stakerPubKeyBytes, err := hex.DecodeString(staker.PublicKey)
		if err != nil {
			return false, err
		}
		stakerHash := computeSelectionHash(stakerPubKeyBytes, drIdBytes)
		totalStakers++
		if bytes.Compare(stakerHash, targetHash) < 0 {
			lowerHashCount++
		}
	}

	if totalStakers == 0 {
		return false, nil
	}

	backupDelayInBlocks := uint64(drConfig.BackupDelayInBlocks)
	replicationFactor := uint64(dr.ReplicationFactor)
	totalNeeded := uint64(0)

	if blocksPassed > backupDelayInBlocks {
		totalNeeded = replicationFactor + ((blocksPassed - 1) / backupDelayInBlocks)
	} else {
		totalNeeded = replicationFactor
	}

	return lowerHashCount < totalNeeded, nil
}
