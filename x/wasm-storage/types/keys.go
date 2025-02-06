package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "wasm-storage"

	// StoreKey defines the primary module store key
	StoreKey = "storage"

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

var (
	// OracleProgramPrefix defines prefix to store oracle programs.
	OracleProgramPrefix = collections.NewPrefix(0)
	// CoreContractRegistryPrefix defines prefix to store address of
	// Core Contract.
	CoreContractRegistryPrefix = collections.NewPrefix(1)
	// ParamsPrefix defines prefix to store parameters of wasm-storage module.
	ParamsPrefix = collections.NewPrefix(2)
)
