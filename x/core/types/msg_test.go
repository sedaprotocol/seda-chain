package types

import (
	"encoding/hex"
	"testing"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	vrf "github.com/sedaprotocol/vrf-go"
	"github.com/stretchr/testify/require"
)

// TestStakeProof tests against a given stake proof.
func TestStakeProof(t *testing.T) {
	chainID := "seda-1-devnet"
	seqNum := uint64(0)
	msg := MsgStake{
		PublicKey: "02bd3aeeea42da249900c97cb63f1647e8c797a432e3b508dcd3e8f70a89b1a622",
		Memo:      "c2VkYXByb3RvY29s",
		Proof:     "030b3a90682d42547987d283027f71e5b434087ae6fd2c46d2cccc870d8b90ca71e0b8331e8467028665341f63a21765c0f4687798aa13b8655fe170f7d7132959e37ae35a2a7e0664a87eb6345c06d507",
	}

	hash, err := msg.ComputeStakeHash(chainID, seqNum)
	require.NoError(t, err)
	publicKey, err := hex.DecodeString(msg.PublicKey)
	require.NoError(t, err)
	proof, err := hex.DecodeString(msg.Proof)
	require.NoError(t, err)

	_, err = vrf.NewK256VRF().Verify(publicKey, proof, hash)
	require.NoError(t, err)
}

// TestProveAndVerifyStakeProof produces a proof and verifies it.
func TestProveAndVerifyStakeProof(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey().Bytes()
	chainID := "seda-1-devnet"
	msg := MsgStake{
		PublicKey: hex.EncodeToString(pubKey),
		Memo:      "VGhlIFNpbmdsZSBVTklYIFNwZWNpZmljYXRpb24gc3VwcG9ydHMgZm9ybWFsIHN0YW5kYXJkcyBkZXZlbG9wZWQgZm9yIGFwcGxpY2F0aW9ucyBwb3J0YWJpbGl0eS4g",
	}
	hash, err := msg.ComputeStakeHash(chainID, 99)
	require.NoError(t, err)

	proof, err := vrf.NewK256VRF().Prove(privKey, hash)
	require.NoError(t, err)

	proof, err = vrf.NewK256VRF().Verify(pubKey, proof, hash)
	require.NoError(t, err)
}
