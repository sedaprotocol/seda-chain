package types

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
		return ErrInvalidValidatorSetTrimPercent
	}
	return nil
}
