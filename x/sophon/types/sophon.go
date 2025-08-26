package types

import "cosmossdk.io/math"

// SophonInputs are all the fields apart from the ID that are needed to
// create a Sophon.
type SophonInputs struct {
	OwnerAddress string
	AdminAddress string
	Address      string
	PublicKey    []byte
	Memo         string
	Balance      math.Int
	UsedCredits  math.Int
}
