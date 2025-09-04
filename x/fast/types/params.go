package types

// DefaultParams returns default fast module parameters.
func DefaultParams() Params {
	return Params{}
}

// ValidateBasic performs basic validation on fast module parameters.
func (p *Params) ValidateBasic() error {
	return nil
}
