package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	DefaultMinFeeUpdateDelay uint32 = 86400 // Roughly 1 week with a ~7 sec block time
	LowestFeeUpdateDelay     uint32 = 1
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
	if p.MinFeeUpdateDelay < LowestFeeUpdateDelay {
		return sdkerrors.ErrInvalidRequest.Wrapf("MinFeeUpdateDelay lower than %d < %d", p.MinFeeUpdateDelay, LowestFeeUpdateDelay)
	}
	return nil
}
