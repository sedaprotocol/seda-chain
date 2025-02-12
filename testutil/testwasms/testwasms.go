package testwasms

import (
	_ "embed"
)

var (
	//go:embed core_contract.wasm
	coreContract []byte

	//go:embed invalid_import.wasm
	invalidImportWasm []byte

	//go:embed random_string_tally.wasm
	randomStringTallyWasm []byte

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

// RandomStringTallyWasm returns a tally wasm binary that concatenates the
// strings found in the 'value' attribute of the reveals. The random
// element from the name only applies to the execution phase, not the tally.
// taken from https://github.com/sedaprotocol/dr-playground 8b40cbc
// Hash: dae38b4ed8c00031a12c1cd506b8f4949b3a314720939f7b5400d1f6b9978337
func RandomStringTallyWasm() []byte {
	return randomStringTallyWasm
}

// InvalidImportsWasm returns a tally wasm binary that has invalid imports.
// Custom build with a non-existent import.
// Hash: 18d962cac6d9f931546f111da8d11b3ed54ccd77b79637b1c49a86ead16edb78
func InvalidImportWasm() []byte {
	return invalidImportWasm
}

func BurnerWasmFileName() string {
	return "burner.wasm"
}

func ReflectWasmFileName() string {
	return "reflect.wasm"
}

func CoreContractWasmFileName() string {
	return "core_contract.wasm"
}
