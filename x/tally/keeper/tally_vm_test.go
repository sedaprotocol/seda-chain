package keeper_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/sedaprotocol/seda-chain/x/tally/keeper"
	"github.com/sedaprotocol/seda-chain/x/tally/keeper/testdata"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

// The name only refers to the oracle program playground name, there is not actually a
// random element in the tally phase.
func TestExecuteTallyProgram_RandomString(t *testing.T) {
	f := initFixture(t)

	tallyProgram := wasmstoragetypes.NewOracleProgram(testdata.RandomStringTallyWasm(), f.Context().BlockTime(), f.Context().BlockHeight(), 1000)
	f.wasmStorageKeeper.OracleProgram.Set(f.Context(), tallyProgram.Hash, tallyProgram)

	gasMeter := types.NewGasMeter(types.DefaultMaxTallyGasLimit, 100, types.DefaultMaxTallyGasLimit, math.NewInt(1), 1)
	vmRes, err := f.tallyKeeper.ExecuteTallyProgram(f.Context(), types.Request{
		TallyProgramID: hex.EncodeToString(tallyProgram.Hash),
		TallyInputs:    base64.StdEncoding.EncodeToString([]byte("hello")),
		PaybackAddress: base64.StdEncoding.EncodeToString([]byte("0x0")),
	}, keeper.FilterResult{
		Outliers: []bool{false, true, false},
	}, []types.RevealBody{
		{
			Reveal:       base64.StdEncoding.EncodeToString([]byte("{\"value\":\"one\"}")),
			ProxyPubKeys: []string{},
			GasUsed:      10,
		},
		{
			Reveal:       base64.StdEncoding.EncodeToString([]byte("{\"value\":\"two\"}")),
			ProxyPubKeys: []string{},
			GasUsed:      10,
		},
		{
			Reveal:       base64.StdEncoding.EncodeToString([]byte("{\"value\":\"three\"}")),
			ProxyPubKeys: []string{},
			GasUsed:      10,
		},
	}, gasMeter)

	require.NoError(t, err)
	require.Equal(t, 0, vmRes.ExitCode)
	require.Equal(t, "one,three", string(vmRes.Result))
}

// TestExecuteTallyProgram_InvalidImports tests that when the tally vm returns a non-zero
// exit code and an empty result, the result is set to the exit_info message. This should
// happen when the tally vm runs into an error outside of the tally program, such as
// an invalid import.
func TestExecuteTallyProgram_InvalidImports(t *testing.T) {
	f := initFixture(t)

	tallyProgram := wasmstoragetypes.NewOracleProgram(testdata.InvalidImportWasm(), f.Context().BlockTime(), f.Context().BlockHeight(), 1000)
	f.wasmStorageKeeper.OracleProgram.Set(f.Context(), tallyProgram.Hash, tallyProgram)

	gasMeter := types.NewGasMeter(types.DefaultMaxTallyGasLimit, 100, types.DefaultMaxTallyGasLimit, math.NewInt(1), 1)
	vmRes, err := f.tallyKeeper.ExecuteTallyProgram(f.Context(), types.Request{
		TallyProgramID: hex.EncodeToString(tallyProgram.Hash),
		TallyInputs:    base64.StdEncoding.EncodeToString([]byte("hello")),
		PaybackAddress: base64.StdEncoding.EncodeToString([]byte("0x0")),
	}, keeper.FilterResult{
		Outliers: []bool{},
	}, []types.RevealBody{}, gasMeter)

	require.NoError(t, err)
	require.NotEqual(t, 0, vmRes.ExitCode)
	require.Contains(t, string(vmRes.Result), "\"seda_v1\".\"this_does_not_exist\"")
}
