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
	// DataRequestPrefix defines prefix to store data request wasms.
	DataRequestPrefix = collections.NewPrefix(0)
	// ExecutorPrefix defines prefix to store executor wasms.
	ExecutorPrefix = collections.NewPrefix(1)
	// WasmExpPrefix defines prefix to track wasm expiration.
	WasmExpPrefix = collections.NewPrefix(2)
	// CoreContractRegistryPrefix defines prefix to store address of
	// Core Contract.
	CoreContractRegistryPrefix = collections.NewPrefix(3)
	// ParamsPrefix defines prefix to store parameters of wasm-storage module.
	ParamsPrefix = collections.NewPrefix(4)
)
