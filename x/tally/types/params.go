package types

import (
	fmt "fmt"

	"cosmossdk.io/math"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	DefaultMaxTallyGasLimit              = 150_000_000_000_000
	DefaultFilterGasCostNone             = 100_000
	DefaultFilterGasCostMultiplierMode   = 100_000
	DefaultFilterGasCostMultiplierStddev = 100_000
	DefaultGasCostBase                   = 1_000_000_000_000
	DefaultGasCostFallback               = 5_000_000_000_000
)

var DefaultBurnRatio = math.LegacyNewDecWithPrec(2, 1)

// DefaultParams returns default tally module parameters.
func DefaultParams() Params {
	return Params{
		MaxTallyGasLimit:              DefaultMaxTallyGasLimit,
		FilterGasCostNone:             DefaultFilterGasCostNone,
		FilterGasCostMultiplierMode:   DefaultFilterGasCostMultiplierMode,
		FilterGasCostMultiplierStdDev: DefaultFilterGasCostMultiplierStddev,
		GasCostBase:                   DefaultGasCostBase,
		GasCostFallback:               DefaultGasCostFallback,
		BurnRatio:                     DefaultBurnRatio,
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
	if p.GasCostFallback <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("commit gas cost must be greater than 0: %d", p.GasCostFallback)
	}
	return validateBurnRatio(p.BurnRatio)
}

func validateBurnRatio(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNil() {
		return fmt.Errorf("burn ratio must be not nil")
	}
	if v.IsNegative() {
		return fmt.Errorf("burn ratio must be positive: %s", v)
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("burn ratio too large: %s", v)
	}
	return nil
}
