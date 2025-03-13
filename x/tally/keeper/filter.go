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
func ExecuteFilter(reveals []types.RevealBody, filterInput string, replicationFactor uint16, params types.Params, gasMeter *types.GasMeter) (FilterResult, error) {
	var res FilterResult
	res.Errors = make([]bool, len(reveals))
	res.Outliers = make([]bool, len(reveals))

	// Determine basic consensus on tuple of (exit_code_success, proxy_pub_keys)
	var maxFreq int
	freq := make(map[string]int, len(reveals))
	for i, reveal := range reveals {
		success := reveal.ExitCode == 0
		res.Errors[i] = !success
		tuple := fmt.Sprintf("%v%v", success, reveal.ProxyPubKeys)
		freq[tuple]++

		if freq[tuple] > maxFreq {
			res.ProxyPubKeys = reveal.ProxyPubKeys
			maxFreq = freq[tuple]
		}
	}
	if maxFreq*3 < int(replicationFactor)*2 {
		res.Consensus, res.Outliers = false, nil
		return res, types.ErrNoBasicConsensus
	}

	filter, err := BuildFilter(filterInput, replicationFactor, params, gasMeter)
	if err != nil {
		res.Consensus, res.Outliers = false, nil
		return res, types.ErrInvalidFilterInput.Wrap(err.Error())
	}

	outliers, consensus := filter.ApplyFilter(reveals, res.Errors)
	switch {
	case countErrors(res.Errors)*3 >= len(reveals)*2:
		res.Consensus, res.Outliers = true, invertErrors(res.Errors)
		return res, types.ErrConsensusInError
	case !consensus:
		res.Consensus, res.Outliers = false, nil
		return res, types.ErrNoConsensus
	default:
		res.Consensus, res.Outliers = true, outliers
		return res, nil
	}
}

// BuildFilter builds a filter based on the requestor-provided input.
func BuildFilter(filterInput string, replicationFactor uint16, params types.Params, gasMeter *types.GasMeter) (types.Filter, error) {
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
		filter, err = types.NewFilterNone(params.FilterGasCostNone, gasMeter)
	case filterTypeMode:
		filter, err = types.NewFilterMode(input, params.FilterGasCostMultiplierMode, replicationFactor, gasMeter)
	case filterTypeStdDev:
		filter, err = types.NewFilterStdDev(input, params.FilterGasCostMultiplierStdDev, replicationFactor, gasMeter)
	default:
		return nil, types.ErrInvalidFilterType
	}
	if err != nil {
		return nil, err
	}
	return filter, nil
}
