package types

import (
	"encoding/hex"
	"testing"

	vrf "github.com/sedaprotocol/vrf-go"
	"github.com/stretchr/testify/require"
)

func TestStakeProof(t *testing.T) {
	chainID := "seda-1-devnet"
	seqNum := uint64(0)
	msg := MsgStake{
		PublicKey: "0284bf7562262bbd6940085748f3be6afa52ae317155181ece31b66351ccffa4b0",
		Memo:      "abcdef",
		Proof:     "03562109e18c4dda4d4a259c066dca85435a64a3e01d3de7da99a5129826c5b295159ab9326d9583b1918a3084984051120d7444fcc6ab3b3366b65ca2ebc7fde9d295e9f46b845c2be382299d333dc139",
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
