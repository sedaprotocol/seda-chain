package types

const (
	EventTypeStoreDataRequestWasm = "store_data_request_wasm"
	EventTypeOverlayWasm          = "store_overlay_wasm"
	EventTypeTallyCompletion      = "tally_completion"
	EventTypeWasmExpiration       = "wasm_expiration"

	AttributeWasmHash      = "wasm_hash"
	AttributeWasmType      = "wasm_type"
	AttributeRequestID     = "request_id"
	AttributeTypeConsensus = "consensus"
)
