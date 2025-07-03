package types

import (
	"fmt"

	"cosmossdk.io/math"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
)

// MarkResultAsFallback marks a DataResult as a fallback result.
// This is used when the request cannot be processed due to an error on our side, as all
// encoding/decoding of values is done by either the contract or the chain.
// It triggers a full refund for the poster.
func MarkResultAsFallback(res *batchingtypes.DataResult, encounteredError error) (err error) {
	gasUsed := math.NewInt(0)

	res.GasUsed = &gasUsed
	res.ExitCode = TallyExitCodeInvalidRequest
	res.Consensus = false
	res.Result = []byte(fmt.Sprintf("unable to process request. error: %s", encounteredError.Error()))

	res.Id, err = res.TryHash()
	if err != nil {
		return err
	}

	return nil
}

// MarkResultAsPaused marks a DataResult as a paused result.
// This is used when the contract is paused and we want to prevent any further processing.
// It triggers a full refund for the poster.
func MarkResultAsPaused(res *batchingtypes.DataResult) (err error) {
	gasUsed := math.NewInt(0)

	res.GasUsed = &gasUsed
	res.ExitCode = TallyExitCodeContractPaused
	res.Consensus = false
	res.Result = []byte("contract is paused")

	res.Id, err = res.TryHash()
	if err != nil {
		return err
	}

	return nil
}
