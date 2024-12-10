package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	DefaultMaxTallyGasLimit = 300_000_000_000_000
)

// DefaultParams returns default tally module parameters.
func DefaultParams() Params {
	return Params{
		MaxTallyGasLimit: DefaultMaxTallyGasLimit,
	}
}

// ValidateBasic performs basic validation on tally module parameters.
func (p *Params) Validate() error {
	if p.MaxTallyGasLimit <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("max tally gas limit must be greater than 0: %d", p.MaxTallyGasLimit)
	}
	return nil
}
