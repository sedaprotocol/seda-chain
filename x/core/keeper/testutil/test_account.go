package testutil

import (
	"encoding/hex"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	vrf "github.com/sedaprotocol/vrf-go"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TestAccount struct {
	name       string
	addr       sdk.AccAddress
	signingKey secp256k1.PrivKey
	fixture    *Fixture
}

func (ta *TestAccount) Name() string {
	return ta.name
}

func (ta *TestAccount) Address() string {
	return ta.addr.String()
}

func (ta *TestAccount) AccAddress() sdk.AccAddress {
	return ta.addr
}

func (ta *TestAccount) PublicKeyHex() string {
	return hex.EncodeToString(ta.signingKey.PubKey().Bytes())
}

func (ta *TestAccount) Balance() math.Int {
	return ta.fixture.BankKeeper.GetBalance(ta.fixture.Context(), ta.addr, BondDenom).Amount
}

func (ta *TestAccount) Prove(hash []byte) string {
	vrf, err := vrf.NewK256VRF().Prove(ta.signingKey.Bytes(), hash)
	require.NoError(ta.fixture.tb, err)
	return hex.EncodeToString(vrf)
}
