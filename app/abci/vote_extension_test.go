package abci

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	cometabci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"

	"github.com/sedaprotocol/seda-chain/app/abci/testutil"
	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/app/utils"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
)

var _ utils.SEDASigner = &mockSigner{}

type mockSigner struct {
	PrivKey secp256k1.PrivKey
}

func (m *mockSigner) Sign(input []byte, index utils.SEDAKeyIndex) ([]byte, error) {
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

	ctrl := gomock.NewController(t)
	mockBatchingKeeper := testutil.NewMockBatchingKeeper(ctrl)
	mockPubKeyKeeper := testutil.NewMockPubKeyKeeper(ctrl)
	mockStakingKeeper := testutil.NewMockStakingKeeper(ctrl)

	// Message: Message for ECDSA signing
	// Private key: 0x79afbf7147841fca72b45a1978dd7669470ba67abbe5c220062924380c9c364b
	// Signature: r=0xb83380f6e1d09411ebf49afd1a95c738686bfb2b0fe2391134f4ae3d6d77b78a, s=0x6c305afcac930a3ea1721c04d8a1a979016baae011319746323a756fbaee1811
	privKeyBytes, err := hex.DecodeString("79afbf7147841fca72b45a1978dd7669470ba67abbe5c220062924380c9c364b")
	require.NoError(t, err)
	privKey := secp256k1.PrivKey(privKeyBytes)
	signer := mockSigner{privKey}

	handler := NewVoteExtensionHandler(
		mockBatchingKeeper,
		mockPubKeyKeeper,
		mockStakingKeeper,
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		&signer,
		nil,
	)

	mockBatch := batchingtypes.Batch{
		BatchId:     []byte("Message for ECDSA signing"),
		BlockHeight: 100, // created from the previous block
	}
	mockBatchingKeeper.EXPECT().GetBatchForHeight(gomock.Any(), gomock.Any()).Return(mockBatch, nil)

	extendVoteHandler := handler.ExtendVoteHandler()
	response, err := extendVoteHandler(sdk.Context{}, &cometabci.RequestExtendVote{
		Height: 101,
	})
	require.NoError(t, err)

	sig1, err := hex.DecodeString("b83380f6e1d09411ebf49afd1a95c738686bfb2b0fe2391134f4ae3d6d77b78a")
	require.NoError(t, err)
	sig2, err := hex.DecodeString("6c305afcac930a3ea1721c04d8a1a979016baae011319746323a756fbaee1811")
	require.NoError(t, err)

	fmt.Println(sig1)
	fmt.Println(sig2)

	require.Equal(t, nil, response.VoteExtension)
}
