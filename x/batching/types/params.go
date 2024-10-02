package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	DefaultValSetTrimPercent uint32 = 95 // 95%
)

// DefaultParams returns default batching module parameters.
func DefaultParams() Params {
	return Params{
		ValidatorSetTrimPercent: DefaultValSetTrimPercent,
	}
}

// Validate validates the batching module parameters.
func (p *Params) Validate() error {
	if p.ValidatorSetTrimPercent > 100 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator set trim percent %d must be between 0 and 100", p.ValidatorSetTrimPercent)
	}
	return nil
}
