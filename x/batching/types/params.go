package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	DefaultNumBatchesToKeep      = 10000
	DefaultMaxBatchPrunePerBlock = 100
)

// DefaultParams returns default batching module parameters.
func DefaultParams() Params {
	return Params{
		NumBatchesToKeep:      DefaultNumBatchesToKeep,
		MaxBatchPrunePerBlock: DefaultMaxBatchPrunePerBlock,
	}
}

// ValidateBasic performs basic validation on batching module parameters.
func (p *Params) Validate() error {
	if p.NumBatchesToKeep <= 3 {
		return sdkerrors.ErrInvalidRequest.Wrapf("num batches to keep must be greater than 3: %d", p.NumBatchesToKeep)
	}
	return nil
}
