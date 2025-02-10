package types

const (
	DefaultMaxWasmSize     int64  = 1024 * 1024    // 1 MB
	DefaultWasmCostPerByte uint64 = 50000000000000 // 0,00005 SEDA
)

// DefaultParams returns default wasm-storage module parameters.
func DefaultParams() Params {
	return Params{
		MaxWasmSize:     DefaultMaxWasmSize,
		WasmCostPerByte: DefaultWasmCostPerByte,
	}
}

// Validate validates the wasm-storage module parameters.
func (p *Params) Validate() error {
	if p.MaxWasmSize <= 0 {
		return ErrInvalidParam.Wrapf("invalid max wasm size %d", p.MaxWasmSize)
	}
	if p.WasmCostPerByte <= 0 {
		return ErrInvalidParam.Wrapf("invalid wasm cost per byte %d", p.WasmCostPerByte)
	}
	return nil
}
