package types

import "cosmossdk.io/errors"

var (
	ErrEmptyRandomnessSeed  = errors.Register(ModuleName, 2, "randomness seed is empty")
	ErrNilVRFSigner         = errors.Register(ModuleName, 3, "vrf signer is nil")
	ErrEmptyPreviousSeed    = errors.Register(ModuleName, 4, "previous seed is empty")
	ErrNewSeedTxNotIncluded = errors.Register(ModuleName, 5, "failed to include a new seed tx in the block")
)
