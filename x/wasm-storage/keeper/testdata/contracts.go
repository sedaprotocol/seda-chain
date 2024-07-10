package testdata

import (
	_ "embed"
)

var (
	//go:embed seda_contract.wasm
	sedaContract []byte

	//go:embed sample_tally.wasm
	sampleTallyWasm []byte

	//go:embed sample_tally_debug.wasm
	sampleTallyDebugWasm []byte
)

func SedaContractWasm() []byte {
	return sedaContract
}

// SampleTallyWasm returns the sample tally wasm binary, whose Keccak256
// hash is 8ade60039246740faa80bf424fc29e79fe13b32087043e213e7bc36620111f6b.
func SampleTallyWasm() []byte {
	return sampleTallyWasm
}

// f49da63e87b982fe8b45eb52c8805ccb9e64cf807989c11ea39b156924d3ac57
func SampleTallyDebugWasm() []byte {
	return sampleTallyDebugWasm
}
