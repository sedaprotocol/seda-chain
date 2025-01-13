package keeper

import (
	"encoding/base64"
	"fmt"

	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

const (
	filterTypeNone   byte = 0x00
	filterTypeMode   byte = 0x01
	filterTypeStdDev byte = 0x02
)

type FilterResult struct {
	Errors       []bool   // i-th item is true if i-th reveal is non-zero exit or corrupt
	Outliers     []bool   // i-th item is non-zero if i-th reveal is an outlier
	Consensus    bool     // whether consensus (either in data or in error) is reached
	ProxyPubKeys []string // data proxy public keys in consensus
	GasUsed      uint64   // gas used by filter
}

// countErrors returns the number of errors in a given error list.
func countErrors(errors []bool) int {
	count := 0
	for _, err := range errors {
		if err {
			count++
		}
	}
	return count
}

// invertErrors returns an inversion of a given error list.
func invertErrors(errors []bool) []bool {
	inverted := make([]bool, len(errors))
	for i, err := range errors {
		inverted[i] = !err
	}
	return inverted
}

// ExecuteFilter builds a filter using the given filter input and applies it to
// the given reveals to determine consensus, proxy public keys in consensus, and
// outliers. It assumes that the reveals are sorted by their keys and that their
// proxy public keys are sorted.
func ExecuteFilter(reveals []types.RevealBody, filterInput string, replicationFactor uint16, params types.Params) (FilterResult, error) {
	filter, err := BuildFilter(filterInput, replicationFactor, params)
	if err != nil {
		return FilterResult{}, types.ErrInvalidFilterInput.Wrap(err.Error())
	}

	var result FilterResult
	result.Errors = make([]bool, len(reveals))
	result.Outliers = make([]bool, len(reveals))
	result.GasUsed = filter.GasCost()

	// Determine basic consensus on tuple of (exit_code_success, proxy_pub_keys)
	var maxFreq int
	freq := make(map[string]int, len(reveals))
	for i, reveal := range reveals {
		success := reveal.ExitCode == 0
		result.Errors[i] = !success
		tuple := fmt.Sprintf("%v%v", success, reveal.ProxyPubKeys)
		freq[tuple]++

		if freq[tuple] > maxFreq {
			result.ProxyPubKeys = reveal.ProxyPubKeys
			maxFreq = freq[tuple]
		}
	}
	if maxFreq*3 < int(replicationFactor)*2 {
		result.Consensus = false
		return result, types.ErrNoBasicConsensus
	}

	outliers, consensus := filter.ApplyFilter(reveals, result.Errors)
	switch {
	case countErrors(result.Errors)*3 > len(reveals)*2:
		result.Consensus = true
		result.Outliers = invertErrors(result.Errors)
		return result, types.ErrConsensusInError
	case !consensus:
		result.Consensus = false
		return result, types.ErrNoConsensus
	default:
		result.Consensus = true
		result.Outliers = outliers
		return result, nil
	}
}

// BuildFilter builds a filter based on the requestor-provided input.
func BuildFilter(filterInput string, replicationFactor uint16, params types.Params) (types.Filter, error) {
	input, err := base64.StdEncoding.DecodeString(filterInput)
	if err != nil {
		return nil, err
	}
	if len(input) == 0 {
		return nil, types.ErrInvalidFilterType
	}

	var filter types.Filter
	switch input[0] {
	case filterTypeNone:
		filter = types.NewFilterNone(params.FilterGasCostNone)
	case filterTypeMode:
		filter, err = types.NewFilterMode(input, params.FilterGasCostMultiplierMode, replicationFactor)
	case filterTypeStdDev:
		filter, err = types.NewFilterStdDev(input, params.FilterGasCostMultiplierStdDev, replicationFactor)
	default:
		return nil, types.ErrInvalidFilterType
	}
	if err != nil {
		return nil, err
	}
	return filter, nil
}
