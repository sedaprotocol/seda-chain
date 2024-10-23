package types

const (
	EventTypeStoreOracleProgram      = "store_oracle_program"
	EventTypeExecutorWasm            = "store_executor_wasm"
	EventTypeOracleProgramExpiration = "oracle_program_expiration"

	AttributeOracleProgramHash = "oracle_program_hash"
	AttributeExecutorWasmHash  = "executor_wasm_hash"
	AttributeSender            = "sender"
)
