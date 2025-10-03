package testutil

import (
	"encoding/hex"

	"github.com/cometbft/cometbft/crypto/secp256k1"
)

type ProxyAccount struct {
	name       string
	privateKey secp256k1.PrivKey
	publicKey  secp256k1.PubKey
	fixture    *Fixture
}

func (pa *ProxyAccount) Name() string {
	return pa.name
}

func (pa *ProxyAccount) PublicKeyHex() string {
	return hex.EncodeToString(pa.publicKey.Bytes())
}
