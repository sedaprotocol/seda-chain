package types

import (
	"fmt"

	"cosmossdk.io/errors"
)

const (
	DefaultMaxWasmSize int64 = 800 * 1024
	DefaultWasmTTL           = 259200 // 21 days
)

// DefaultParams returns default wasm-storage module parameters.
func DefaultParams() Params {
	return Params{
		MaxWasmSize: DefaultMaxWasmSize,
		WasmTTL:     DefaultWasmTTL,
	}
}

// ValidateBasic performs basic validation on wasm-storage
// module parameters.
func (p *Params) ValidateBasic() error {
	if p.WasmTTL < 2 {
		return errors.Wrapf(ErrInvalidParam, "WasmTTL %d < 2", p.WasmTTL)
	}
	return validateMaxWasmSize(p.MaxWasmSize)
}

func validateMaxWasmSize(i int64) error {
	if i == 0 {
		return fmt.Errorf("invalid max Wasm size: %d", i)
	}
	return nil
}
