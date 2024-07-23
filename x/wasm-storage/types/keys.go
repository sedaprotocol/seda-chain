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
	// DataRequestPrefix defines prefix to store Data Request Wasm binaries.
	DataRequestPrefix = collections.NewPrefix(0)
	// OverlayPrefix defines prefix to store Overlay Wasm binaries.
	OverlayPrefix = collections.NewPrefix(1)
	// WasmExpPrefix defines prefix to track wasm expiration.
	WasmExpPrefix = collections.NewPrefix(2)
	// CoreContractRegistryPrefix defines prefix to store address of
	// Core Contract.
	CoreContractRegistryPrefix = collections.NewPrefix(3)
	// ParamsPrefix defines prefix to store parameters of wasm-storage module.
	ParamsPrefix = collections.NewPrefix(4)
)
