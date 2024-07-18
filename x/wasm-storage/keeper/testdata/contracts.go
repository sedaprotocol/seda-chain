package testdata

import (
	_ "embed"
)

var (
	//go:embed seda_contract.wasm
	sedaContract []byte

	//go:embed sample_tally.wasm
	sampleTallyWasm []byte
)

func SedaContractWasm() []byte {
	return sedaContract
}

// SampleTallyWasm returns the sample tally wasm binary, whose Keccak256
// hash is 8ade60039246740faa80bf424fc29e79fe13b32087043e213e7bc36620111f6b.
func SampleTallyWasm() []byte {
	return sampleTallyWasm
}
