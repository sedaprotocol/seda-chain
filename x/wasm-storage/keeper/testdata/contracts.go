package testdata

import (
	_ "embed"
)

var (
	//go:embed seda_contract.wasm
	sedaContract []byte
)

func SedaContractWasm() []byte {
	return sedaContract
}
