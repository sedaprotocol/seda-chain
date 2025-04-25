package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	DefaultActivationBlockDelay       = 11520 // roughly 1 day
	DefaultActivationThresholdPercent = 80
)

// DefaultParams returns default pubkey module parameters.
func DefaultParams() Params {
	return Params{
		ActivationBlockDelay:       DefaultActivationBlockDelay,
		ActivationThresholdPercent: DefaultActivationThresholdPercent,
	}
}

// ValidateBasic performs basic validation on pubkey module parameters.
func (p *Params) Validate() error {
	if p.ActivationBlockDelay < 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("ActivationBlockDelay should not be negative: %d", p.ActivationBlockDelay)
	}
	if p.ActivationThresholdPercent < 67 || p.ActivationThresholdPercent > 100 {
		return sdkerrors.ErrInvalidRequest.Wrapf("ActivationThresholdPercent should be between 67 and 100: %d", p.ActivationThresholdPercent)
	}
	return nil
}
