package types

import (
	"cosmossdk.io/math"
)

var (
	DefaultMinimumStake = math.NewInt(1000000000000000000)
)

const (
	DefaultCommitTimeoutInBlocks       uint32 = 50
	DefaultRevealTimeoutInBlocks       uint32 = 5
	DefaultBackupDelayInBlocks         uint32 = 5
	DefaultDrRevealSizeLimitInBytes    uint32 = 24000 // 24 KB
	DefaultExecInputLimitInBytes       uint32 = 2048  // 2 KB
	DefaultTallyInputLimitInBytes      uint32 = 512   // 512 B
	DefaultConsensusFilterLimitInBytes uint32 = 512   // 512 B
	DefaultMemoLimitInBytes            uint32 = 512   // 512 B
	DefaultPaybackAddressLimitInBytes  uint32 = 128   // 128 B
	DefaultSedaPayloadLimitInBytes     uint32 = 512   // 512 B
	DefaultAllowlistEnabled            bool   = true
)

// DefaultParams returns default core module parameters.
func DefaultParams() Params {
	return Params{
		DataRequestConfig: DataRequestConfig{
			CommitTimeoutInBlocks:       DefaultCommitTimeoutInBlocks,
			RevealTimeoutInBlocks:       DefaultRevealTimeoutInBlocks,
			BackupDelayInBlocks:         DefaultBackupDelayInBlocks,
			DrRevealSizeLimitInBytes:    DefaultDrRevealSizeLimitInBytes,
			ExecInputLimitInBytes:       DefaultExecInputLimitInBytes,
			TallyInputLimitInBytes:      DefaultTallyInputLimitInBytes,
			ConsensusFilterLimitInBytes: DefaultConsensusFilterLimitInBytes,
			MemoLimitInBytes:            DefaultMemoLimitInBytes,
			PaybackAddressLimitInBytes:  DefaultPaybackAddressLimitInBytes,
			SedaPayloadLimitInBytes:     DefaultSedaPayloadLimitInBytes,
		},
		StakingConfig: StakingConfig{
			MinimumStake:     DefaultMinimumStake,
			AllowlistEnabled: DefaultAllowlistEnabled,
		},
	}
}

// ValidateBasic performs basic validation on core module parameters.
func (p *Params) Validate() error {
	// TODO
	return nil
}
