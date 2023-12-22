package app

import "strings"

// The last arguments can contain custom message handlers, and custom query handlers,
// if we want to allow any custom callbacks
// See https://github.com/CosmWasm/cosmwasm/blob/main/docs/CAPABILITIES-BUILT-IN.md
var wasmCapabilities = []string{
	"iterator",
	"staking",
	"stargate",
	"cosmwasm_1_1",
	"cosmwasm_1_2",
}

func GetWasmCapabilities() string {
	return strings.Join(wasmCapabilities, ",")
}
