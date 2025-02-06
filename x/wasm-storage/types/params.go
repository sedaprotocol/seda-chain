package types

const (
	DefaultMaxWasmSize      int64 = 800 * 1024
	DefaultWasmTTL                = 259200 // 21 days
	DefaultUploadMultiplier       = 200
)

// DefaultParams returns default wasm-storage module parameters.
func DefaultParams() Params {
	return Params{
		MaxWasmSize:      DefaultMaxWasmSize,
		WasmTTL:          DefaultWasmTTL,
		UploadMultiplier: DefaultUploadMultiplier,
	}
}

// Validate validates the wasm-storage module parameters.
func (p *Params) Validate() error {
	if p.WasmTTL < 2 {
		return ErrInvalidParam.Wrapf("WasmTTL %d < 2", p.WasmTTL)
	}
	if p.MaxWasmSize <= 0 {
		return ErrInvalidParam.Wrapf("invalid max wasm size %d", p.MaxWasmSize)
	}
	if p.UploadMultiplier <= 0 {
		return ErrInvalidParam.Wrapf("invalid upload multiplier %d", p.UploadMultiplier)
	}
	return nil
}
