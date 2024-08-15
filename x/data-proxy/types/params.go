package types

import (
	"cosmossdk.io/errors"
)

const (
	DefaultMinFeeUpdateDelay uint32 = 86400 // Roughly 1 week with a ~7 sec block time
)

// DefaultParams returns default wasm-storage module parameters.
func DefaultParams() Params {
	return Params{
		MinFeeUpdateDelay: DefaultMinFeeUpdateDelay,
	}
}

// ValidateBasic performs basic validation on wasm-storage
// module parameters.
func (p *Params) Validate() error {
	if p.MinFeeUpdateDelay < 1 {
		return errors.Wrapf(ErrInvalidParam, "MinFeeUpdateDelay %d < 1", p.MinFeeUpdateDelay)
	}
	return nil
}
