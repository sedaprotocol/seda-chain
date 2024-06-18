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

func SampleTallyWasm() []byte {
	return sampleTallyWasm
}
