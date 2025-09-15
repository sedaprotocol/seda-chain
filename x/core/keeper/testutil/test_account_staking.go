package testutil

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	vrf "github.com/sedaprotocol/vrf-go"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (ta *TestAccount) Stake(amountSeda int64, memo string) (*types.MsgStakeResponse, error) {
	bigAmountSeda := math.NewInt(amountSeda)
	bigAmount := bigAmountSeda.Mul(math.NewInt(1_000_000_000_000_000_000))
	stake := sdk.NewCoin(BondDenom, bigAmount)

	memoBase64 := base64.StdEncoding.EncodeToString([]byte(memo))
	msg := &types.MsgStake{
		Sender:    ta.Address(),
		PublicKey: ta.PublicKeyHex(),
		Memo:      memoBase64,
		Proof:     "",
		Stake:     stake,
	}
	hash, err := msg.MsgHash(ta.fixture.ChainID, ta.GetSequence())
	require.NoError(ta.fixture.tb, err)
	proof, err := vrf.NewK256VRF().Prove(ta.signingKey.Bytes(), hash)
	require.NoError(ta.fixture.tb, err)
	msg.Proof = hex.EncodeToString(proof)

	return ta.fixture.CoreMsgServer.Stake(ta.fixture.Context(), msg)
}

func (ta *TestAccount) Unstake() (*types.MsgUnstakeResponse, error) {
	msg := &types.MsgUnstake{
		Sender:    ta.Address(),
		PublicKey: ta.PublicKeyHex(),
	}
	hash, err := msg.MsgHash(ta.fixture.ChainID, ta.GetSequence())
	require.NoError(ta.fixture.tb, err)
	proof, err := vrf.NewK256VRF().Prove(ta.signingKey.Bytes(), hash)
	require.NoError(ta.fixture.tb, err)
	msg.Proof = hex.EncodeToString(proof)

	return ta.fixture.CoreMsgServer.Unstake(ta.fixture.Context(), msg)
}

func (ta *TestAccount) Withdraw(to *TestAccount) (*types.MsgWithdrawResponse, error) {
	msg := &types.MsgWithdraw{
		Sender:    ta.Address(),
		PublicKey: ta.PublicKeyHex(),
	}

	if to != nil {
		msg.WithdrawAddress = to.Address()
	} else {
		msg.WithdrawAddress = ta.Address()
	}

	hash, err := msg.MsgHash(ta.fixture.ChainID, ta.GetSequence())
	require.NoError(ta.fixture.tb, err)
	proof, err := vrf.NewK256VRF().Prove(ta.signingKey.Bytes(), hash)
	require.NoError(ta.fixture.tb, err)
	msg.Proof = hex.EncodeToString(proof)

	return ta.fixture.CoreMsgServer.Withdraw(ta.fixture.Context(), msg)
}
