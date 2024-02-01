package types

import (
	fmt "fmt"
)

const DefaultMaxWasmSize uint64 = 800 * 1024

// DefaultParams returns default wasm-storage module parameters.
func DefaultParams() Params {
	return Params{
		MaxWasmSize: DefaultMaxWasmSize,
	}
}

// ValidateBasic performs basic validation on wasm-storage
// module parameters.
func (p Params) ValidateBasic() error {
	err := validateMaxWasmSize(p.MaxWasmSize)
	return err
}

func validateMaxWasmSize(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("invalid max Wasm size: %d", v)
	}
	return nil
}
