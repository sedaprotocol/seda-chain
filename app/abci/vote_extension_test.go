package abci

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/skip-mev/slinky/abci/strategies/codec"
	"github.com/skip-mev/slinky/abci/testutils"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/sha3"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cometabci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/abci/testutil"
	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/app/utils"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

func TestABCIHandlers(t *testing.T) {
	// Set up SEDA signer.
	tmpDir := t.TempDir()

	valConsAddr := sdk.ConsAddress([]byte("testval"))
	valValAddr := sdk.ValAddress(simtestutil.CreateRandomAccounts(1)[0])

	sedaPubKeys, err := utils.GenerateSEDAKeys(valValAddr, tmpDir, "", false)
	require.NoError(t, err)
	secp256k1PubKey := sedaPubKeys[utils.SEDAKeyIndexSecp256k1].PubKey
	ethAddr, err := utils.PubKeyToEthAddress(secp256k1PubKey)
	require.NoError(t, err)

	signer, err := utils.LoadSEDASigner(filepath.Join(tmpDir, utils.SEDAKeyFileName), true)
	require.NoError(t, err)

	// Configure address prefixes.
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(params.Bech32PrefixAccAddr, params.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(params.Bech32PrefixValAddr, params.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(params.Bech32PrefixConsAddr, params.Bech32PrefixConsPub)
	cfg.Seal()

	buf := &bytes.Buffer{}
	logger := log.NewLogger(buf, log.LevelOption(zerolog.DebugLevel))

	// Mock configurations (Batch 100 is created at height H)
	ctrl := gomock.NewController(t)
	mockBatchingKeeper := testutil.NewMockBatchingKeeper(ctrl)
	mockPubKeyKeeper := testutil.NewMockPubKeyKeeper(ctrl)
	mockStakingKeeper := testutil.NewMockStakingKeeper(ctrl)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte("Message for ECDSA signing"))
	batchID := hasher.Sum(nil)
	mockBatch := batchingtypes.Batch{
		BatchNumber: 100,
		BatchId:     batchID,
		BlockHeight: 100,
	}

	mockBatchingKeeper.EXPECT().GetBatchForHeight(gomock.Any(), mockBatch.BlockHeight).Return(mockBatch, nil).AnyTimes()
	mockBatchingKeeper.EXPECT().GetValidatorTreeEntry(gomock.Any(), mockBatch.BatchNumber-1, valValAddr).Return(
		batchingtypes.ValidatorTreeEntry{
			EthAddress: ethAddr,
		}, nil).AnyTimes()
	mockPubKeyKeeper.EXPECT().GetValidatorKeys(gomock.Any(), valValAddr.String()).Return(pubkeytypes.ValidatorPubKeys{}, nil)

	// Construct the handler.
	defaultProposalHandler := baseapp.NewDefaultProposalHandler(mempool.NoOpMempool{}, nil)
	handler := NewHandlers(
		defaultProposalHandler.PrepareProposalHandler(),
		defaultProposalHandler.ProcessProposalHandler(),
		NewNoOpVoteExtensionValidator(),
		mockBatchingKeeper,
		mockPubKeyKeeper,
		mockStakingKeeper,
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		signer,
		logger,
	)
	extendVoteHandler := handler.ExtendVoteHandler()
	verifyVoteHandler := handler.VerifyVoteExtensionHandler()
	prepareProposalHandler := handler.PrepareProposalHandler()
	preBlockerHandler := handler.PreBlocker()

	// ExtendVote at H+1
	ctx := sdk.Context{}.
		WithBlockHeight(101).
		WithConsensusParams(cmtproto.ConsensusParams{
			Abci: &cmtproto.ABCIParams{
				VoteExtensionsEnableHeight: 100,
			},
		})
	evRes, err := extendVoteHandler(ctx, &cometabci.RequestExtendVote{
		Height: 101,
	})
	require.NoError(t, err)

	// Recover and verify public key
	sigPubKey, err := crypto.Ecrecover(mockBatch.BatchId, evRes.VoteExtension)
	require.NoError(t, err)
	require.Equal(t, secp256k1PubKey, sigPubKey)

	mockStakingKeeper.EXPECT().GetValidatorByConsAddr(gomock.Any(), valConsAddr).Return(
		stakingtypes.Validator{
			OperatorAddress: valValAddr.String(),
		}, nil,
	).AnyTimes()

	// VerifyVoteExtension at H+1
	vvRes, err := verifyVoteHandler(ctx, &cometabci.RequestVerifyVoteExtension{
		Height:           101,
		VoteExtension:    evRes.VoteExtension,
		ValidatorAddress: valConsAddr,
	})
	require.NoError(t, err)
	require.Equal(t, cometabci.ResponseVerifyVoteExtension_ACCEPT, vvRes.Status)

	// PrepareProposal at H+2
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	veCodec := codec.NewDefaultVoteExtensionCodec()
	emptyVote, err := testutils.CreateExtendedVoteInfo(valConsAddr, map[uint64][]byte{}, veCodec)
	require.NoError(t, err)
	emptyVote.VoteExtension = evRes.VoteExtension

	extendedCommit := cometabci.ExtendedCommitInfo{
		Round: 1,
		Votes: []cometabci.ExtendedVoteInfo{
			emptyVote,
		},
	}

	prepareRes, err := prepareProposalHandler(ctx, &cometabci.RequestPrepareProposal{
		LocalLastCommit: extendedCommit,
		MaxTxBytes:      22020096,
		Height:          102,
	})
	require.NoError(t, err)

	// Ensure last commit is injected here.
	var injected abcitypes.ExtendedCommitInfo
	err = json.Unmarshal(prepareRes.Txs[0], &injected)
	require.NoError(t, err)
	require.Equal(t, extendedCommit, injected)

	// ProcessProposal at H+2
	processRes, err := handler.ProcessProposalHandler()(ctx, &cometabci.RequestProcessProposal{
		ProposedLastCommit: cometabci.CommitInfo{
			Round: 1,
			Votes: nil,
		},
		Txs:    prepareRes.Txs,
		Height: 102,
	})
	require.NoError(t, err)
	require.Equal(t, cometabci.ResponseProcessProposal_ACCEPT, processRes.Status)

	// PreBlocker at H+2
	mockBatchingKeeper.EXPECT().SetBatchSigSecp256k1(gomock.Any(), mockBatch.BatchNumber, valValAddr, evRes.VoteExtension).Return(nil)

	_, err = preBlockerHandler(ctx, &cometabci.RequestFinalizeBlock{
		Txs:    prepareRes.Txs,
		Height: 102,
	})
	require.NoError(t, err)
}
