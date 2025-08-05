package keeper

import (
	"encoding/hex"
	"testing"

	"github.com/sedaprotocol/seda-chain/x/core/types"
	vrf "github.com/sedaprotocol/vrf-go"
	"github.com/stretchr/testify/require"
)

func TestStakeProof(t *testing.T) {
	chainID := "seda-1-devnet"
	seqNum := uint64(0)
	msg := types.MsgStake{
		PublicKey: "03d92f44157c939284bb101dccea8a2fc95f71ecfd35b44573a76173e3c25c67a9",
		// Memo:      "",
		Proof: "032c74385c590d76e1a6e15364f515f0ae38ba61077c276dcf6aea4a810a36e4988a32cccfd9b08c8ab74f3e4e6dbb6f8e600364432bb166361018f45b817b350b30ae352b7131ab267dffcd643057c483",
	}

	hash, err := msg.ComputeStakeHash("seda1nr9t0fe333uql6hh4k8h9qs8mzstjr4qsea3y5smyrd2ptqpt85sev3d8l", chainID, seqNum)
	require.NoError(t, err)
	publicKey, err := hex.DecodeString(msg.PublicKey)
	require.NoError(t, err)
	proof, err := hex.DecodeString(msg.Proof)
	require.NoError(t, err)

	_, err = vrf.NewK256VRF().Verify(publicKey, proof, hash)
	require.NoError(t, err)
}
