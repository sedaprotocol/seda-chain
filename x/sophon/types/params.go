package types

// DefaultParams returns default sophon module parameters.
func DefaultParams() Params {
	return Params{}
}

// ValidateBasic performs basic validation on sophon module parameters.
func (p *Params) ValidateBasic() error {
	return nil
}
