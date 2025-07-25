package types

const (
	EventTypeTallyCompletion = "tally_completion"
	EventTypeGasMeter        = "gas_calculation"

	AttributeDataRequestID     = "dr_id"
	AttributeDataRequestHeight = "dr_height"
	AttributeDataResultID      = "id"
	AttributeTypeConsensus     = "consensus"
	AttributeTallyVMStdOut     = "tally_vm_stdout"
	AttributeTallyVMStdErr     = "tally_vm_stderr"
	AttributeExecGasUsed       = "exec_gas_used"
	AttributeTallyGasUsed      = "tally_gas_used"
	AttributeTallyExitCode     = "exit_code"
	AttributeProxyPubKeys      = "proxy_public_keys"
	AttributeTallyGas          = "tally_gas"
	AttributeDataProxyGas      = "data_proxy_gas"
	AttributeExecutorGas       = "executor_reward_gas"
	AttributeReducedPayout     = "reduced_payout"
	AttributeReducedPayoutBurn = "reduced_payout_burn"
)
