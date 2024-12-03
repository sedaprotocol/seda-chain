package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	DefaultActivationLag = 25
)

// DefaultParams returns default pubkey module parameters.
func DefaultParams() Params {
	return Params{
		ActivationLag: DefaultActivationLag,
	}
}

// ValidateBasic performs basic validation on pubkey module parameters.
func (p *Params) Validate() error {
	if p.ActivationLag < 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("ActivationLag should not be negative: %d", p.ActivationLag)
	}
	return nil
}
