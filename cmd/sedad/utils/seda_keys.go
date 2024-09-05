package utils

import (
	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/secp256k1"
)

const (
	SEDAKeysIndexSecp256k1 = 0
)

// SEDAKeysGenerators is a map from SEDA Key Index to the
// corresponding private key generator function.
var SEDAKeysGenerators = map[uint32]PrivKeyGenerator{
	SEDAKeysIndexSecp256k1: secp256k1GenPrivKey,
}

type PrivKeyGenerator func() cmtcrypto.PrivKey

func secp256k1GenPrivKey() cmtcrypto.PrivKey {
	return secp256k1.GenPrivKey()
}
