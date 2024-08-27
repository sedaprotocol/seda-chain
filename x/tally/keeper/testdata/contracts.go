package testdata

import (
	_ "embed"
)

var (
	//go:embed core_contract.wasm
	coreContract []byte

	//go:embed sample_tally.wasm
	sampleTallyWasm []byte

	//go:embed sample_tally_2.wasm
	sampleTallyWasm2 []byte
)

func CoreContractWasm() []byte {
	return coreContract
}

// SampleTallyWasm returns a sample tally wasm binary whose Keccak256
// hash is 8ade60039246740faa80bf424fc29e79fe13b32087043e213e7bc36620111f6b.
func SampleTallyWasm() []byte {
	return sampleTallyWasm
}

// SampleTallyWasm2 returns a tally wasm binary that prints all the
// tally VM environment variables.
// Hash: 5f3b31bff28c64a143119ee6389d62e38767672daace9c36db54fa2d18e9f391
func SampleTallyWasm2() []byte {
	return sampleTallyWasm2
}
