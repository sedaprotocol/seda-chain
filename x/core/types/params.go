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
	DefaultMaxResultSize               = 1024
	DefaultMaxTallyGasLimit            = 50_000_000_000_000
	DefaultFilterGasCostNone           = 100_000
	DefaultFilterGasCostMultiplierMode = 100_000
	DefaultFilterGasCostMultiplierMAD  = 100_000
	DefaultGasCostBase                 = 1_000_000_000_000
	DefaultExecutionGasCostFallback    = 5_000_000_000_000
	DefaultMaxTalliesPerBlock          = 100
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
			GasCostBase:                 DefaultGasCostBase,
			ExecutionGasCostFallback:    DefaultExecutionGasCostFallback,
			BurnRatio:                   DefaultBurnRatio,
			MaxTalliesPerBlock:          DefaultMaxTalliesPerBlock,
		},
	}
}

// ValidateBasic performs basic validation on core module parameters.
func (p *Params) Validate() error {
	err := p.TallyConfig.Validate()
	if err != nil {
		return err
	}

	// TODO: Add validation for other configs

	return nil
}

// ValidateBasic performs basic validation on tally module parameters.
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
	if tc.GasCostBase <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("base gas cost must be greater than 0: %d", tc.GasCostBase)
	}
	if tc.ExecutionGasCostFallback <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("execution gas cost fallback must be greater than 0: %d", tc.ExecutionGasCostFallback)
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
