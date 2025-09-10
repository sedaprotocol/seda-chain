package testutil

import (
	"encoding/hex"

	"github.com/cometbft/cometbft/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TestAccount struct {
	name       string
	addr       sdk.AccAddress
	signingKey secp256k1.PrivKey
	fixture    *Fixture
	Sequence   uint64
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

func (ta *TestAccount) GetSequence() uint64 {
	// TODO: should use the query to get the actual sequence number
	current := ta.Sequence
	ta.Sequence++
	return current
}
