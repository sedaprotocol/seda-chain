package testdata

import (
	_ "embed"
)

var (
	//go:embed data_requests.wasm
	dataRequestsContract []byte
)

func DataRequestsContractWasm() []byte {
	return dataRequestsContract
}
