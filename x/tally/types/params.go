package types

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
	return nil
}
