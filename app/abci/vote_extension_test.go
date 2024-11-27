package abci

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/sha3"

	cometabci "github.com/cometbft/cometbft/abci/types"

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

func TestExtendVerifyVoteHandlers(t *testing.T) {
	// Set up SEDA signer.
	tmpDir := t.TempDir()

	valAddr := sdk.ValAddress(simtestutil.CreateRandomAccounts(1)[0])
	pubKeys, err := utils.GenerateSEDAKeys(valAddr, tmpDir)
	require.NoError(t, err)
	signer, err := utils.LoadSEDASigner(filepath.Join(tmpDir, utils.SEDAKeyFileName))
	require.NoError(t, err)

	secp256k1PubKey := pubKeys[utils.SEDAKeyIndexSecp256k1].PubKey

	ethAddr, err := utils.PubKeyToEthAddress(secp256k1PubKey)
	require.NoError(t, err)

	// Configure address prefixes.
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(params.Bech32PrefixAccAddr, params.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(params.Bech32PrefixValAddr, params.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(params.Bech32PrefixConsAddr, params.Bech32PrefixConsPub)
	cfg.Seal()

	buf := &bytes.Buffer{}
	logger := log.NewLogger(buf, log.LevelOption(zerolog.DebugLevel))

	ctx := sdk.Context{}.WithBlockHeight(101)

	// Mock configurations
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
		BlockHeight: 100, // created from the previous block
	}

	mockBatchingKeeper.EXPECT().GetBatchForHeight(gomock.Any(), int64(100)).Return(mockBatch, nil).AnyTimes()
	mockBatchingKeeper.EXPECT().GetValidatorTreeEntry(gomock.Any(), uint64(99), valAddr).Return(
		batchingtypes.ValidatorTreeEntry{
			EthAddress: ethAddr,
		}, nil).AnyTimes()
	mockPubKeyKeeper.EXPECT().GetValidatorKeys(gomock.Any(), valAddr.String()).Return(pubkeytypes.ValidatorPubKeys{}, nil)

	// Construct the handler and execute it.
	handler := NewHandlers(
		baseapp.NewDefaultProposalHandler(mempool.NoOpMempool{}, nil),
		mockBatchingKeeper,
		mockPubKeyKeeper,
		mockStakingKeeper,
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		signer,
		logger,
	)
	extendVoteHandler := handler.ExtendVoteHandler()
	verifyVoteHandler := handler.VerifyVoteExtensionHandler()

	evRes, err := extendVoteHandler(ctx, &cometabci.RequestExtendVote{
		Height: 101,
	})
	require.NoError(t, err)

	// Recover and verify public key
	sigPubKey, err := crypto.Ecrecover(mockBatch.BatchId, evRes.VoteExtension)
	require.NoError(t, err)
	require.Equal(t, secp256k1PubKey, sigPubKey)

	testVal := sdk.ConsAddress([]byte("testval"))
	mockStakingKeeper.EXPECT().GetValidatorByConsAddr(gomock.Any(), testVal).Return(
		stakingtypes.Validator{
			OperatorAddress: valAddr.String(),
		}, nil,
	)

	vvRes, err := verifyVoteHandler(ctx, &cometabci.RequestVerifyVoteExtension{
		Height:           101,
		VoteExtension:    evRes.VoteExtension,
		ValidatorAddress: testVal,
	})
	require.NoError(t, err)
	require.Equal(t, cometabci.ResponseVerifyVoteExtension_ACCEPT, vvRes.Status)
}
