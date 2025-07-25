package types

import (
	"cosmossdk.io/math"
)

var (
	DefaultMinimumStake = math.NewInt(1000000000000000000)
)

const (
	DefaultAllowlistEnabled = false
)

// DefaultParams returns default core module parameters.
func DefaultParams() Params {
	return Params{
		DataRequestConfig: DataRequestConfig{},
		StakingConfig: StakingConfig{
			MinimumStake:     DefaultMinimumStake,
			AllowlistEnabled: DefaultAllowlistEnabled,
		},
	}
}

// ValidateBasic performs basic validation on core module parameters.
func (p *Params) Validate() error {
	// TODO
	return nil
}
