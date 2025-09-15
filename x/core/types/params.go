package types

import (
	"fmt"

	"cosmossdk.io/math"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	DefaultMinimumStake = math.NewInt(1000000000000000000)
	DefaultBurnRatio    = math.LegacyNewDecWithPrec(2, 1)
)

const (
	// StakingConfig
	DefaultAllowlistEnabled bool = true

	// DataRequestConfig
	DefaultCommitTimeoutInBlocks       uint32 = 50
	DefaultRevealTimeoutInBlocks       uint32 = 5
	DefaultBackupDelayInBlocks         uint32 = 5
	DefaultDrRevealSizeLimitInBytes    uint32 = 24000 // 24 KB
	DefaultExecInputLimitInBytes       uint32 = 2048  // 2 KB
	DefaultTallyInputLimitInBytes      uint32 = 512   // 512 B
	DefaultConsensusFilterLimitInBytes uint32 = 512   // 512 B
	DefaultMemoLimitInBytes            uint32 = 512   // 512 B
	DefaultPaybackAddressLimitInBytes  uint32 = 128   // 128 B
	DefaultSEDAPayloadLimitInBytes     uint32 = 512   // 512 B

	// TallyConfig
	DefaultMaxResultSize               uint32 = 1024
	DefaultMaxTallyGasLimit            uint64 = 50_000_000_000_000
	DefaultFilterGasCostNone           uint64 = 100_000
	DefaultFilterGasCostMultiplierMode uint64 = 100_000
	DefaultFilterGasCostMultiplierMAD  uint64 = 100_000
	DefaultBaseGasCost                 uint64 = 1_000_000_000_000
	DefaultExecutionGasCostFallback    uint64 = 5_000_000_000_000
	DefaultMaxTalliesPerBlock          uint32 = 100
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
			SEDAPayloadLimitInBytes:     DefaultSEDAPayloadLimitInBytes,
		},
		StakingConfig: StakingConfig{
			MinimumStake:     DefaultMinimumStake,
			AllowlistEnabled: DefaultAllowlistEnabled,
		},
		TallyConfig: TallyConfig{
			MaxResultSize:               DefaultMaxResultSize,
			MaxTallyGasLimit:            DefaultMaxTallyGasLimit,
			FilterGasCostNone:           DefaultFilterGasCostNone,
			FilterGasCostMultiplierMode: DefaultFilterGasCostMultiplierMode,
			FilterGasCostMultiplierMAD:  DefaultFilterGasCostMultiplierMAD,
			BaseGasCost:                 DefaultBaseGasCost,
			ExecutionGasCostFallback:    DefaultExecutionGasCostFallback,
			BurnRatio:                   DefaultBurnRatio,
			MaxTalliesPerBlock:          DefaultMaxTalliesPerBlock,
		},
	}
}

// ValidateBasic performs basic validation on core module parameters.
func (p *Params) Validate() error {
	err := p.DataRequestConfig.Validate()
	if err != nil {
		return err
	}
	err = p.StakingConfig.Validate()
	if err != nil {
		return err
	}
	err = p.TallyConfig.Validate()
	if err != nil {
		return err
	}
	return nil
}

func (dc *DataRequestConfig) Validate() error {
	if dc.CommitTimeoutInBlocks <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("commit timeout must be greater than 0: %d blocks", dc.CommitTimeoutInBlocks)
	}
	if dc.RevealTimeoutInBlocks <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("reveal timeout must be greater than 0: %d blocks", dc.RevealTimeoutInBlocks)
	}
	if dc.BackupDelayInBlocks <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("backup delay must be greater than 0: %d blocks", dc.BackupDelayInBlocks)
	}
	if dc.DrRevealSizeLimitInBytes <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("reveal size limit must be greater than 0: %d bytes", dc.DrRevealSizeLimitInBytes)
	}
	if dc.ExecInputLimitInBytes <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("exec input limit must be greater than 0: %d bytes", dc.ExecInputLimitInBytes)
	}
	if dc.TallyInputLimitInBytes <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("tally input limit must be greater than 0: %d bytes", dc.TallyInputLimitInBytes)
	}
	if dc.ConsensusFilterLimitInBytes <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("consensus filter limit must be greater than 0: %d bytes", dc.ConsensusFilterLimitInBytes)
	}
	if dc.MemoLimitInBytes <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("memo limit must be greater than 0: %d bytes", dc.MemoLimitInBytes)
	}
	if dc.PaybackAddressLimitInBytes <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("payback address limit must be greater than 0: %d bytes", dc.PaybackAddressLimitInBytes)
	}
	if dc.SEDAPayloadLimitInBytes <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("SEDA payload limit must be greater than 0: %d bytes", dc.SEDAPayloadLimitInBytes)
	}
	return nil
}

func (sc *StakingConfig) Validate() error {
	if !sc.MinimumStake.IsPositive() {
		return sdkerrors.ErrInvalidRequest.Wrapf("minimum stake must be positive: %s", sc.MinimumStake)
	}
	return nil
}

func (tc *TallyConfig) Validate() error {
	if tc.MaxResultSize <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("max result size must be greater than 0: %d", tc.MaxResultSize)
	}
	if tc.MaxTallyGasLimit <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("max tally gas limit must be greater than 0: %d", tc.MaxTallyGasLimit)
	}
	if tc.FilterGasCostNone <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("filter gas cost (none) must be greater than 0: %d", tc.FilterGasCostNone)
	}
	if tc.FilterGasCostMultiplierMode <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("filter gas cost (mode) must be greater than 0: %d", tc.FilterGasCostMultiplierMode)
	}
	if tc.FilterGasCostMultiplierMAD <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("filter gas cost (MAD) must be greater than 0: %d", tc.FilterGasCostMultiplierMAD)
	}
	if tc.BaseGasCost <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("base gas cost must be greater than 0: %d", tc.BaseGasCost)
	}
	if tc.ExecutionGasCostFallback <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("execution gas cost fallback must be greater than 0: %d", tc.ExecutionGasCostFallback)
	}
	if tc.MaxTalliesPerBlock <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("max tallies per block must be greater than 0: %d", tc.MaxTalliesPerBlock)
	}
	return validateBurnRatio(tc.BurnRatio)
}

func validateBurnRatio(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNil() {
		return fmt.Errorf("burn ratio must be not nil")
	}
	if v.IsNegative() {
		return fmt.Errorf("burn ratio must be positive: %s", v)
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("burn ratio too large: %s", v)
	}
	return nil
}
