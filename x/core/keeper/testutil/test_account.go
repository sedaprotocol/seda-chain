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
