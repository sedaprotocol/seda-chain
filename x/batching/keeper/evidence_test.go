package keeper_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"

	sdkstakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/x/batching/keeper"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

func TestHandleEvidence(t *testing.T) {
	f := initFixture(t)

	valAddrs, privKeys, validators := generateFirstBatch(t, f, 10)

	fraudAddr := valAddrs[0]
	fraudPrivKey := privKeys[0]
	fraudulentValidator := validators[0]

	// Keep track of all validator tokens before the evidence is handled
	validatorTokensBefore := make(map[string]sdkmath.Int, len(validators))
	for _, validator := range validators {
		validatorTokensBefore[validator.OperatorAddress] = validator.GetTokens()
	}

	// Store the legitimate batch for which the validator will be double signing
	doubleSignBatchNumber := uint64(1)
	doubleSignBlockHeight := int64(4)

	batchToDoubleSign := types.Batch{
		BatchId:     []byte("batch2"),
		BatchNumber: doubleSignBatchNumber,
		BlockHeight: doubleSignBlockHeight,
	}
	err := f.batchingKeeper.SetNewBatch(f.Context(), batchToDoubleSign, types.DataResultTreeEntries{}, []types.ValidatorTreeEntry{})
	require.NoError(t, err)

	f.stakingKeeper.SetHistoricalInfo(f.Context(), doubleSignBlockHeight, &sdkstakingtypes.HistoricalInfo{
		Valset: validators,
	})

	// Create evidence for a fraudulent batch
	evidence := &types.BatchDoubleSign{
		BatchNumber:         doubleSignBatchNumber,
		BlockHeight:         doubleSignBlockHeight,
		OperatorAddress:     fraudAddr.String(),
		DataResultRoot:      "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
		ValidatorRoot:       "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
		ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
		ProvingSchemeIndex:  0,
	}

	// Sign fraudulent batch
	fraudulentBatchID, err := evidence.GetBatchID()
	require.NoError(t, err)
	signature, err := crypto.Sign(fraudulentBatchID, fraudPrivKey)
	require.NoError(t, err)

	evidence.Signature = hex.EncodeToString(signature)

	// Test handling evidence
	handler := keeper.NewBatchDoubleSignHandler(f.batchingKeeper)
	err = handler(f.Context().WithBlockHeight(doubleSignBlockHeight), evidence)
	require.NoError(t, err)

	// Verify validator was slashed, jailed, and tombstoned
	fraudulentValidatorAfter, err := f.stakingKeeper.GetValidator(f.Context(), fraudAddr)
	require.NoError(t, err)
	require.True(t, fraudulentValidatorAfter.IsJailed())
	tokensAfter := fraudulentValidatorAfter.GetTokens()
	require.True(t, tokensAfter.LT(fraudulentValidator.GetTokens()))
	consAddr, err := fraudulentValidatorAfter.GetConsAddr()
	require.NoError(t, err)
	require.True(t, f.slashingKeeper.IsTombstoned(f.Context(), consAddr))

	// Verify all other validators are unaffected
	for _, valAddr := range valAddrs[1:] {
		validatorAfter, err := f.stakingKeeper.GetValidator(f.Context(), valAddr)
		require.NoError(t, err)
		require.False(t, validatorAfter.IsJailed())
		require.Equal(t, validatorAfter.GetTokens(), validatorTokensBefore[validatorAfter.OperatorAddress])
	}
}

func TestHandleEvidence_DifferentBlockHeight(t *testing.T) {
	f := initFixture(t)
	valAddrs, privKeys, validators := generateFirstBatch(t, f, 1)

	fraudAddr := valAddrs[0]
	fraudPrivKey := privKeys[0]

	// Keep track of all validator tokens before the evidence is handled
	validatorTokensBefore := make(map[string]sdkmath.Int, len(validators))
	for _, validator := range validators {
		validatorTokensBefore[validator.OperatorAddress] = validator.GetTokens()
	}

	// Store the legitimate batch for which the validator will be double signing
	doubleSignBatchNumber := uint64(1)
	doubleSignBlockHeight := int64(4)

	batchToDoubleSign := types.Batch{
		BatchId:     []byte("batch2"),
		BatchNumber: doubleSignBatchNumber,
		BlockHeight: doubleSignBlockHeight,
	}
	err := f.batchingKeeper.SetNewBatch(f.Context(), batchToDoubleSign, types.DataResultTreeEntries{}, []types.ValidatorTreeEntry{})
	require.NoError(t, err)

	f.stakingKeeper.SetHistoricalInfo(f.Context(), doubleSignBlockHeight, &sdkstakingtypes.HistoricalInfo{
		Valset: validators,
	})

	// Create evidence for a fraudulent batch
	evidence := &types.BatchDoubleSign{
		BatchNumber:         doubleSignBatchNumber,
		BlockHeight:         doubleSignBlockHeight + 100,
		OperatorAddress:     fraudAddr.String(),
		DataResultRoot:      "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
		ValidatorRoot:       "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
		ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
		ProvingSchemeIndex:  0,
	}

	// Sign fraudulent batch
	fraudulentBatchID, err := evidence.GetBatchID()
	require.NoError(t, err)
	signature, err := crypto.Sign(fraudulentBatchID, fraudPrivKey)
	require.NoError(t, err)

	evidence.Signature = hex.EncodeToString(signature)

	// Test handling evidence
	handler := keeper.NewBatchDoubleSignHandler(f.batchingKeeper)
	err = handler(f.Context().WithBlockHeight(doubleSignBlockHeight), evidence)
	require.NoError(t, err)

	fraudulentValidator, err := f.stakingKeeper.GetValidator(f.Context(), fraudAddr)
	require.NoError(t, err)
	require.True(t, fraudulentValidator.IsJailed())
	consAddr, err := fraudulentValidator.GetConsAddr()
	require.NoError(t, err)
	require.True(t, f.slashingKeeper.IsTombstoned(f.Context(), consAddr))
}

func TestHandleEvidence_FutureBatch(t *testing.T) {
	f := initFixture(t)
	_, _, _ = generateFirstBatch(t, f, 1)

	evidence := &types.BatchDoubleSign{
		BatchNumber: 999, // Non-existent batch
	}

	handler := keeper.NewBatchDoubleSignHandler(f.batchingKeeper)

	err := handler(f.Context(), evidence)
	require.ErrorIs(t, err, collections.ErrNotFound)
}

func TestHandleEvidence_InvalidBatchID(t *testing.T) {
	f := initFixture(t)
	_, _, _ = generateFirstBatch(t, f, 1)

	evidence := &types.BatchDoubleSign{
		BatchNumber:         0,
		BlockHeight:         1,
		DataResultRoot:      "x027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
		ValidatorRoot:       "y2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
		ProvingMetadataHash: "z000000000000000000000000000000000000000000000000000000000000000",
	}

	handler := keeper.NewBatchDoubleSignHandler(f.batchingKeeper)
	err := handler(f.Context(), evidence)
	require.ErrorIs(t, err, hex.InvalidByteError('y'))
}

func TestHandleEvidence_LegitBatchID(t *testing.T) {
	f := initFixture(t)

	valAddrs, privKeys, _ := generateFirstBatch(t, f, 1)

	notFraudAddr := valAddrs[0]
	notFraudPrivKey := privKeys[0]

	// Create evidence for a legitimate batch
	evidence := &types.BatchDoubleSign{
		BatchNumber:         1,
		BlockHeight:         2,
		OperatorAddress:     notFraudAddr.String(),
		DataResultRoot:      "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
		ValidatorRoot:       "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
		ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
		ProvingSchemeIndex:  0,
	}

	legitBatchID, err := evidence.GetBatchID()
	require.NoError(t, err)

	err = f.batchingKeeper.SetNewBatch(f.Context(), types.Batch{
		BatchId:     legitBatchID,
		BatchNumber: 1,
		BlockHeight: 2,
	}, types.DataResultTreeEntries{}, []types.ValidatorTreeEntry{})
	require.NoError(t, err)

	// Sign legitimate batch
	signature, err := crypto.Sign(legitBatchID, notFraudPrivKey)
	require.NoError(t, err)
	evidence.Signature = hex.EncodeToString(signature)

	handler := keeper.NewBatchDoubleSignHandler(f.batchingKeeper)
	err = handler(f.Context().WithBlockHeight(2), evidence)
	require.Equal(t, err, fmt.Errorf("batch IDs are the same"))
}

func TestHandleEvidence_InvalidSignature(t *testing.T) {
	f := initFixture(t)
	_, _, _ = generateFirstBatch(t, f, 1)

	testcases := []struct {
		name        string
		signature   string
		errorString string
	}{
		{name: "invalid signature length", signature: "", errorString: "invalid signature length"},
		{name: "invalid signature hex", signature: "x1234567890", errorString: "invalid byte"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			evidence := &types.BatchDoubleSign{
				BatchNumber:         0,
				BlockHeight:         1,
				DataResultRoot:      "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
				ValidatorRoot:       "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
				ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
				ProvingSchemeIndex:  0,
				Signature:           tc.signature,
			}

			handler := keeper.NewBatchDoubleSignHandler(f.batchingKeeper)
			err := handler(f.Context(), evidence)
			require.ErrorContains(t, err, tc.errorString)
		})
	}
}

func TestHandleEvidence_DifferentPrivateKey(t *testing.T) {
	f := initFixture(t)

	valAddrs, privKeys, validators := generateFirstBatch(t, f, 2)

	fraudAddr := valAddrs[0]
	fraudPrivKey := privKeys[1]

	// Store the legitimate batch for which the validator will be double signing
	doubleSignBatchNumber := uint64(1)
	doubleSignBlockHeight := int64(4)

	batchToDoubleSign := types.Batch{
		BatchId:     []byte("batch2"),
		BatchNumber: doubleSignBatchNumber,
		BlockHeight: doubleSignBlockHeight,
	}
	err := f.batchingKeeper.SetNewBatch(f.Context(), batchToDoubleSign, types.DataResultTreeEntries{}, []types.ValidatorTreeEntry{})
	require.NoError(t, err)

	f.stakingKeeper.SetHistoricalInfo(f.Context(), doubleSignBlockHeight, &sdkstakingtypes.HistoricalInfo{
		Valset: validators,
	})

	// Create evidence for a fraudulent batch
	evidence := &types.BatchDoubleSign{
		BatchNumber:         doubleSignBatchNumber,
		BlockHeight:         doubleSignBlockHeight,
		OperatorAddress:     fraudAddr.String(),
		DataResultRoot:      "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
		ValidatorRoot:       "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
		ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
		ProvingSchemeIndex:  0,
	}

	// Sign fraudulent batch with a different private key
	fraudulentBatchID, err := evidence.GetBatchID()
	require.NoError(t, err)
	signature, err := crypto.Sign(fraudulentBatchID, fraudPrivKey)
	require.NoError(t, err)

	evidence.Signature = hex.EncodeToString(signature)

	// Test handling evidence
	handler := keeper.NewBatchDoubleSignHandler(f.batchingKeeper)
	err = handler(f.Context().WithBlockHeight(doubleSignBlockHeight), evidence)
	require.ErrorContains(t, err, "recovered address does not match validator entry")
}

func TestHandleEvidence_StaleEvidence(t *testing.T) {
	f := initFixture(t)

	valAddrs, privKeys, validators := generateFirstBatch(t, f, 1)

	fraudAddr := valAddrs[0]
	fraudPrivKey := privKeys[0]

	// Store the legitimate batch for which the validator will be double signing
	doubleSignBatchNumber := uint64(1)
	doubleSignBlockHeight := int64(4)

	batchToDoubleSign := types.Batch{
		BatchId:     []byte("batch2"),
		BatchNumber: doubleSignBatchNumber,
		BlockHeight: doubleSignBlockHeight,
	}
	err := f.batchingKeeper.SetNewBatch(f.Context(), batchToDoubleSign, types.DataResultTreeEntries{}, []types.ValidatorTreeEntry{})
	require.NoError(t, err)

	f.stakingKeeper.SetHistoricalInfo(f.Context(), doubleSignBlockHeight, &sdkstakingtypes.HistoricalInfo{
		Valset: validators,
	})

	// Create evidence for a fraudulent batch
	evidence := &types.BatchDoubleSign{
		BatchNumber:         doubleSignBatchNumber,
		BlockHeight:         doubleSignBlockHeight,
		OperatorAddress:     fraudAddr.String(),
		DataResultRoot:      "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
		ValidatorRoot:       "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
		ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
		ProvingSchemeIndex:  0,
	}

	// Sign fraudulent batch
	fraudulentBatchID, err := evidence.GetBatchID()
	require.NoError(t, err)
	signature, err := crypto.Sign(fraudulentBatchID, fraudPrivKey)
	require.NoError(t, err)

	evidence.Signature = hex.EncodeToString(signature)

	consParams := f.Context().ConsensusParams()
	consParams.Evidence = &cmtproto.EvidenceParams{
		MaxAgeNumBlocks: 1,
	}

	// Test handling evidence
	handler := keeper.NewBatchDoubleSignHandler(f.batchingKeeper)
	err = handler(f.Context().WithBlockHeight(doubleSignBlockHeight+10).WithConsensusParams(consParams), evidence)
	require.NoError(t, err)

	// Verify all validators are unaffected
	for _, valAddr := range valAddrs {
		validator, err := f.stakingKeeper.GetValidator(f.Context(), valAddr)
		require.NoError(t, err)
		require.False(t, validator.IsJailed())
	}

	// Test the same input with a block height that is within the max age
	err = handler(f.Context().WithBlockHeight(doubleSignBlockHeight).WithConsensusParams(consParams), evidence)
	require.NoError(t, err)

	fraudulentValidator, err := f.stakingKeeper.GetValidator(f.Context(), fraudAddr)
	require.NoError(t, err)
	require.True(t, fraudulentValidator.IsJailed())

	// Verify all other validators are unaffected
	for _, valAddr := range valAddrs[1:] {
		validator, err := f.stakingKeeper.GetValidator(f.Context(), valAddr)
		require.NoError(t, err)
		require.False(t, validator.IsJailed())
	}
}

func TestHandleEvidence_TombstonedValidator(t *testing.T) {
	f := initFixture(t)

	valAddrs, privKeys, validators := generateFirstBatch(t, f, 1)

	fraudAddr := valAddrs[0]
	fraudPrivKey := privKeys[0]
	fraudulentValidator := validators[0]

	// Store the legitimate batch for which the validator will be double signing
	doubleSignBatchNumber := uint64(1)
	doubleSignBlockHeight := int64(4)

	batchToDoubleSign := types.Batch{
		BatchId:     []byte("batch2"),
		BatchNumber: doubleSignBatchNumber,
		BlockHeight: doubleSignBlockHeight,
	}
	err := f.batchingKeeper.SetNewBatch(f.Context(), batchToDoubleSign, types.DataResultTreeEntries{}, []types.ValidatorTreeEntry{})
	require.NoError(t, err)

	f.stakingKeeper.SetHistoricalInfo(f.Context(), doubleSignBlockHeight, &sdkstakingtypes.HistoricalInfo{
		Valset: validators,
	})

	// Create evidence for a fraudulent batch
	evidence := &types.BatchDoubleSign{
		BatchNumber:         doubleSignBatchNumber,
		BlockHeight:         doubleSignBlockHeight,
		OperatorAddress:     fraudAddr.String(),
		DataResultRoot:      "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
		ValidatorRoot:       "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
		ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
		ProvingSchemeIndex:  0,
	}

	// Sign fraudulent batch
	fraudulentBatchID, err := evidence.GetBatchID()
	require.NoError(t, err)
	signature, err := crypto.Sign(fraudulentBatchID, fraudPrivKey)
	require.NoError(t, err)

	evidence.Signature = hex.EncodeToString(signature)

	// Tombstone the validator before handling evidence
	consAddr, err := fraudulentValidator.GetConsAddr()
	require.NoError(t, err)
	f.slashingKeeper.Tombstone(f.Context(), consAddr)

	// Test handling evidence
	handler := keeper.NewBatchDoubleSignHandler(f.batchingKeeper)
	err = handler(f.Context().WithBlockHeight(doubleSignBlockHeight), evidence)
	require.NoError(t, err)

	// Verify the validator was not slashed
	fraudulentValidatorAfter, err := f.stakingKeeper.GetValidator(f.Context(), fraudAddr)
	require.NoError(t, err)
	require.Equal(t, fraudulentValidatorAfter.GetTokens(), fraudulentValidator.GetTokens())
}
