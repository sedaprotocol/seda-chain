package abci

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	cometabci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/abci/testutil"
	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/app/utils"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
)

var _ utils.SEDASigner = &mockSigner{}

type mockSigner struct {
	PrivKey secp256k1.PrivKey
}

func (m *mockSigner) Sign(input []byte, _ utils.SEDAKeyIndex) ([]byte, error) {
	signature, err := m.PrivKey.Sign(input)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func TestExtendVoteHandler(t *testing.T) {
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

	mockBatch := batchingtypes.Batch{
		BatchId:     []byte("Message for ECDSA signing"),
		BlockHeight: 100, // created from the previous block
	}

	privKeyBytes, err := hex.DecodeString("79afbf7147841fca72b45a1978dd7669470ba67abbe5c220062924380c9c364b")
	require.NoError(t, err)
	privKey := secp256k1.PrivKey{Key: privKeyBytes}
	signer := mockSigner{privKey}
	expSig, err := privKey.Sign(mockBatch.BatchId)
	require.NoError(t, err)

	mockBatchingKeeper.EXPECT().GetBatchForHeight(gomock.Any(), int64(100)).Return(mockBatch, nil).AnyTimes()

	testVal := sdk.ConsAddress([]byte("testval"))
	mockStakingKeeper.EXPECT().GetValidatorByConsAddr(gomock.Any(), testVal).Return(
		stakingtypes.Validator{
			OperatorAddress: "sedavaloper1ucv5709wlf9jn84ynyjzyzeavwvurmdydtn3px",
		}, nil,
	)

	mockPubKeyKeeper.EXPECT().GetValidatorKeyAtIndex(
		gomock.Any(),
		[]byte{230, 25, 79, 60, 174, 250, 75, 41, 158, 164, 153, 36, 34, 11, 61, 99, 153, 193, 237, 164},
		utils.Secp256k1,
	).Return(privKey.PubKey(), nil)

	// Construct the handler and execute it.
	handler := NewHandlers(
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
	require.Equal(t, expSig, evRes.VoteExtension)

	vvRes, err := verifyVoteHandler(ctx, &cometabci.RequestVerifyVoteExtension{
		Height:           101,
		VoteExtension:    evRes.VoteExtension,
		ValidatorAddress: testVal,
	})
	require.NoError(t, err)
	require.Equal(t, cometabci.ResponseVerifyVoteExtension_ACCEPT, vvRes.Status)
}
