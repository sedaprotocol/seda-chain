package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	DefaultMaxTallyGasLimit              = 150_000_000_000_000
	DefaultFilterGasCostNone             = 100_000
	DefaultFilterGasCostMultiplierMode   = 100_000
	DefaultFilterGasCostMultiplierStddev = 100_000
	DefaultGasCostBase                   = 1_000_000_000_000
	DefaultGasCostCommit                 = 5_000_000_000_000
)

// DefaultParams returns default tally module parameters.
func DefaultParams() Params {
	return Params{
		MaxTallyGasLimit:              DefaultMaxTallyGasLimit,
		FilterGasCostNone:             DefaultFilterGasCostNone,
		FilterGasCostMultiplierMode:   DefaultFilterGasCostMultiplierMode,
		FilterGasCostMultiplierStdDev: DefaultFilterGasCostMultiplierStddev,
		GasCostBase:                   DefaultGasCostBase,
		GasCostCommit:                 DefaultGasCostCommit,
	}
}

// ValidateBasic performs basic validation on tally module parameters.
func (p *Params) Validate() error {
	if p.MaxTallyGasLimit <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("max tally gas limit must be greater than 0: %d", p.MaxTallyGasLimit)
	}
	if p.FilterGasCostNone <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("filter gas cost (none) must be greater than 0: %d", p.FilterGasCostNone)
	}
	if p.FilterGasCostMultiplierMode <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("filter gas cost (mode) must be greater than 0: %d", p.FilterGasCostMultiplierMode)
	}
	if p.FilterGasCostMultiplierStdDev <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("filter gas cost (std dev) must be greater than 0: %d", p.FilterGasCostMultiplierStdDev)
	}
	if p.GasCostBase <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("base gas cost must be greater than 0: %d", p.GasCostBase)
	}
	if p.GasCostCommit <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("commit gas cost must be greater than 0: %d", p.GasCostCommit)
	}
	return nil
}
