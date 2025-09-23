package keeper

import (
	"fmt"

	"cosmossdk.io/math"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// MarkResultAsFallback marks a DataResult as a fallback result.
// This is used when the request cannot be processed due to an error on our side, as all
// encoding/decoding of values is done by either the contract or the chain.
// It triggers a full refund for the poster.
func MarkResultAsFallback(dataResult *batchingtypes.DataResult, tallyResult *TallyResult, encounteredError error) (err error) {
	gasUsed := math.NewInt(0)

	dataResult.GasUsed = &gasUsed
	dataResult.ExitCode = types.TallyExitCodeInvalidRequest
	dataResult.Consensus = false
	dataResult.Result = []byte(fmt.Sprintf("unable to process request. error: %s", encounteredError.Error()))

	dataResult.Id, err = dataResult.TryHash()
	if err != nil {
		return err
	}

	tallyResult.FilterResult = FilterResult{Error: types.ErrFilterDidNotRun}
	return nil
}

// MarkResultAsPaused marks a DataResult as a paused result.
// This is used when the contract is paused and we want to prevent any further processing.
// It triggers a full refund for the poster.
func MarkResultAsPaused(dataResult *batchingtypes.DataResult, tallyResult *TallyResult) (err error) {
	gasUsed := math.NewInt(0)

	dataResult.GasUsed = &gasUsed
	dataResult.ExitCode = types.TallyExitCodeContractPaused
	dataResult.Consensus = false
	dataResult.Result = []byte("contract is paused")

	dataResult.Id, err = dataResult.TryHash()
	if err != nil {
		return err
	}

	tallyResult.FilterResult = FilterResult{Error: types.ErrFilterDidNotRun}
	return nil
}
