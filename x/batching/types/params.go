package types

const (
	DefaultMaxBatchPrunePerBlock            = 100
	DefaultMaxLegacyDataResultPrunePerBlock = 1000
)

// DefaultParams returns default batching module parameters.
func DefaultParams() Params {
	return Params{
		MaxBatchPrunePerBlock:            DefaultMaxBatchPrunePerBlock,
		MaxLegacyDataResultPrunePerBlock: DefaultMaxLegacyDataResultPrunePerBlock,
	}
}

// ValidateBasic performs basic validation on batching module parameters.
func (p *Params) Validate() error {
	return nil
}
