package abci

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/sha3"

	cometabci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/abci/testutil"
	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/app/utils"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
)

var _ utils.SEDASigner = &mockSigner{}

type mockSigner struct {
	PrivKey *ecdsa.PrivateKey
}

func (m *mockSigner) Sign(input []byte, _ utils.SEDAKeyIndex) ([]byte, error) {
	signature, err := crypto.Sign(input, m.PrivKey)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func TestExtendVerifyVoteHandlers(t *testing.T) {
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
		BatchId:     batchID,
		BlockHeight: 100, // created from the previous block
	}

	privKey, err := crypto.HexToECDSA("79afbf7147841fca72b45a1978dd7669470ba67abbe5c220062924380c9c364b")
	require.NoError(t, err)
	pubKey := elliptic.Marshal(privKey.PublicKey, privKey.PublicKey.X, privKey.PublicKey.Y)

	signer := mockSigner{privKey}

	mockBatchingKeeper.EXPECT().GetBatchForHeight(gomock.Any(), int64(100)).Return(mockBatch, nil).AnyTimes()

	// Construct the handler and execute it.
	handler := NewHandlers(
		baseapp.NewDefaultProposalHandler(mempool.NoOpMempool{}, nil),
		mockBatchingKeeper,
		mockPubKeyKeeper,
		mockStakingKeeper,
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		logger,
	)
	handler.SetSEDASigner(&signer)
	extendVoteHandler := handler.ExtendVoteHandler()
	verifyVoteHandler := handler.VerifyVoteExtensionHandler()

	evRes, err := extendVoteHandler(ctx, &cometabci.RequestExtendVote{
		Height: 101,
	})
	require.NoError(t, err)

	// Recover and verify public key
	sigPubKey, err := crypto.Ecrecover(mockBatch.BatchId, evRes.VoteExtension)
	require.NoError(t, err)
	require.Equal(t, pubKey, sigPubKey)

	testVal := sdk.ConsAddress([]byte("testval"))
	mockStakingKeeper.EXPECT().GetValidatorByConsAddr(gomock.Any(), testVal).Return(
		stakingtypes.Validator{
			OperatorAddress: "sedavaloper1ucv5709wlf9jn84ynyjzyzeavwvurmdydtn3px",
		}, nil,
	)
	mockPubKeyKeeper.EXPECT().GetValidatorKeyAtIndex(
		gomock.Any(),
		[]byte{230, 25, 79, 60, 174, 250, 75, 41, 158, 164, 153, 36, 34, 11, 61, 99, 153, 193, 237, 164},
		utils.SEDAKeyIndexSecp256k1,
	).Return(pubKey, nil)

	vvRes, err := verifyVoteHandler(ctx, &cometabci.RequestVerifyVoteExtension{
		Height:           101,
		VoteExtension:    evRes.VoteExtension,
		ValidatorAddress: testVal,
	})
	require.NoError(t, err)
	require.Equal(t, cometabci.ResponseVerifyVoteExtension_ACCEPT, vvRes.Status)
}
