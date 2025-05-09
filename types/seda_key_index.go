package types

import (
	"fmt"
)

// SEDAKeyIndex enumerates the SEDA key indices.
type SEDAKeyIndex uint32

const (
	SEDAKeyIndexSecp256k1 SEDAKeyIndex = iota
)

// SEDA domain separators
const (
	SEDASeparatorDataResult byte = iota
	SEDASeparatorSecp256k1
)

func (i SEDAKeyIndex) String() string {
	switch i {
	case SEDAKeyIndexSecp256k1:
		return "SEDA_KEY_INDEX_SECP256K1"
	default:
		return fmt.Sprintf("unknown(%d)", i)
	}
}
