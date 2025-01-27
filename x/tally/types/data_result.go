package types

import (
	"fmt"

	"cosmossdk.io/math"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
)

// MarkResultAsFallback marks a DataResult as a fallback result.
// This is used when the request cannot be processed due to an error on our side, as all
// encoding/decoding of values is done by either the contract or the chain.
func MarkResultAsFallback(res *batchingtypes.DataResult, err error, exitCode int) {
	gasUsed := math.NewInt(0)

	res.GasUsed = &gasUsed
	//nolint:gosec // G115: We shouldn't get negative exit codes.
	res.ExitCode = uint32(exitCode)
	res.Consensus = false
	res.Result = []byte(fmt.Sprintf("unable to process request. error: %s", err.Error()))
}
