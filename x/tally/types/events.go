package types

const (
	EventTypeTallyCompletion         = "tally_completion"
	EventTypeDataProxyReward         = "data_proxy_reward"
	EventTypeExecutorRewardUniform   = "executor_reward_uniform"
	EventTypeExecutorRewardDivergent = "executor_reward_divergent"
	EventTypeExecutorRewardCommit    = "executor_reward_commit"
	EventTypeTallyGasBurn            = "tally_gas_burn"
	EventTypeBaseFeeBurn             = "base_fee_burn"

	AttributeDataRequestID = "dr_id"
	AttributeDataResultID  = "id"
	AttributeTypeConsensus = "consensus"
	AttributeTallyVMStdOut = "tally_vm_stdout"
	AttributeTallyVMStdErr = "tally_vm_stderr"
	AttributeExecGasUsed   = "exec_gas_used"
	AttributeTallyGasUsed  = "tally_gas_used"
	AttributeTallyExitCode = "exit_code"
	AttributeProxyPubKeys  = "proxy_public_keys"
	AttributeProxyPubKey   = "proxy_public_key"
	AttributeExecutor      = "executor"
)
