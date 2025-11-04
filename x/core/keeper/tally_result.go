package keeper

import (
	"sort"

	"cosmossdk.io/math"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// TallyResult is used to track results of tally process that are not covered
// by DataResult.
type TallyResult struct {
	// ID is the data request ID in hex.
	ID string
	// Height is the height at which the data request was posted.
	Height            uint64
	ReplicationFactor uint16
	Reveals           []types.Reveal
	GasMeter          *types.GasMeter
	GasReports        []uint64
	FilterResult      FilterResult
	StdOut            []string
	StdErr            []string
	ExecGasUsed       uint64
	TallyGasUsed      uint64
}

// MeterExecutorGasUniform computes and records the gas consumption of executors
// when their gas reports are uniformly at "gasReport". If a non-nil outliers
// slice is provided, no gas consumption will be recorded for the executors
// specified as outliers.
func (t TallyResult) MeterExecutorGasUniform() {
	executorGasReport := t.GasMeter.CorrectExecGasReportWithProxyGas(t.GasReports[0])
	gasUsed := min(executorGasReport, t.GasMeter.RemainingExecGas()/uint64(t.ReplicationFactor))
	for i, r := range t.Reveals {
		if t.FilterResult.Outliers != nil && t.FilterResult.Outliers[i] {
			continue
		}
		t.GasMeter.ConsumeExecGasForExecutor(r.Executor, gasUsed)
	}
}

// MeterExecutorGasDivergent computes and records the gas consumption of executors
// when their gas reports are divergent. If a non-nil outliers slice is provided,
// no gas consumption will be recorded for the executors specified as outliers.
func (t TallyResult) MeterExecutorGasDivergent() {
	var lowestReport uint64
	var lowestReporterIndex int
	adjGasReports := make([]uint64, len(t.GasReports))
	for i, gasReport := range t.GasReports {
		executorGasReport := t.GasMeter.CorrectExecGasReportWithProxyGas(gasReport)
		adjGasReports[i] = min(executorGasReport, t.GasMeter.RemainingExecGas()/uint64(t.ReplicationFactor))
		if i == 0 || adjGasReports[i] < lowestReport {
			lowestReporterIndex = i
			lowestReport = adjGasReports[i]
		}
	}
	medianGasUsed := median(adjGasReports)
	totalGasUsed := math.NewIntFromUint64(medianGasUsed*uint64(t.ReplicationFactor-1) + min(lowestReport*2, medianGasUsed))
	totalShares := math.NewIntFromUint64(medianGasUsed * uint64(t.ReplicationFactor-1)).Add(math.NewIntFromUint64(lowestReport * 2))
	var lowestGasUsed, regGasUsed uint64
	if totalShares.IsZero() {
		lowestGasUsed = 0
		regGasUsed = 0
	} else {
		lowestGasUsed = math.NewIntFromUint64(lowestReport * 2).Mul(totalGasUsed).Quo(totalShares).Uint64()
		regGasUsed = math.NewIntFromUint64(medianGasUsed).Mul(totalGasUsed).Quo(totalShares).Uint64()
	}
	for i, r := range t.Reveals {
		if t.FilterResult.Outliers != nil && t.FilterResult.Outliers[i] {
			continue
		}
		gasUsed := regGasUsed
		if i == lowestReporterIndex {
			gasUsed = lowestGasUsed
		}
		t.GasMeter.ConsumeExecGasForExecutor(r.Executor, gasUsed)
	}
}

func median(arr []uint64) uint64 {
	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})
	n := len(arr)
	if n%2 == 0 {
		return (arr[n/2-1] + arr[n/2]) / 2
	}
	return arr[n/2]
}
