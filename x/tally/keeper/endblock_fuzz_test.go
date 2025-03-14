package keeper_test

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func getRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return bytes
}

type fuzzCommitReveal struct {
	requestHeight uint64
	requestMemo   string // base64-encoded
	reveal        string // base64-encoded
	exitCode      byte
	gasUsed       uint64
	salt          string
	proxyPubKeys  []string
}

func FuzzEndBlock(f *testing.F) {
	fixture := initFixture(f)

	defaultParams := types.DefaultParams()
	err := fixture.tallyKeeper.SetParams(fixture.Context(), defaultParams)
	require.NoError(f, err)

	f.Fuzz(func(t *testing.T, reveal, proxyPubKey []byte, exitCode byte, requestHeight, gasUsed uint64) {
		sim := fuzzCommitReveal{
			requestHeight: requestHeight,
			requestMemo:   base64.StdEncoding.EncodeToString(getRandomBytes(10)),
			reveal:        base64.StdEncoding.EncodeToString(reveal),
			exitCode:      exitCode,
			gasUsed:       gasUsed,
			proxyPubKeys:  []string{hex.EncodeToString(proxyPubKey)},
		}

		drID, _ := fixture.fuzzCommitRevealDataRequest(t, sim, 4, false)

		err = fixture.tallyKeeper.EndBlock(fixture.Context())
		require.NoError(t, err)
		require.NotContains(t, fixture.logBuf.String(), "ERR")

		_, err := fixture.batchingKeeper.GetLatestDataResult(fixture.Context(), drID)
		require.NoError(t, err)
	})
}

// commitRevealDataRequest simulates stakers committing and revealing
// for a data request. It returns the data request ID.
func (f *fixture) fuzzCommitRevealDataRequest(t *testing.T, fuzz fuzzCommitReveal, replicationFactor int, timeout bool) (string, []staker) {
	stakers := f.addStakers(t, 5)

	// Upload data request and tally oracle programs.
	execProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm(), f.Context().BlockTime())
	err := f.wasmStorageKeeper.OracleProgram.Set(f.Context(), execProgram.Hash, execProgram)
	require.NoError(t, err)

	tallyProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm2(), f.Context().BlockTime())
	err = f.wasmStorageKeeper.OracleProgram.Set(f.Context(), tallyProgram.Hash, tallyProgram)
	require.NoError(t, err)

	// Post a data request.
	resJSON, err := f.contractKeeper.Execute(
		f.Context(),
		f.coreContractAddr,
		f.deployer,
		postDataRequestMsg(execProgram.Hash, tallyProgram.Hash, fuzz.requestMemo, replicationFactor),
		sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1003000000000000000))),
	)
	require.NoError(t, err)

	type PostDataRequestResponse struct {
		DrID   string `json:"dr_id"`
		Height uint64 `json:"height"`
	}
	var res PostDataRequestResponse
	err = json.Unmarshal(resJSON, &res)
	require.NoError(t, err)
	drID := res.DrID

	// The stakers commit and reveal.
	revealBody := types.RevealBody{
		RequestID:          drID,
		RequestBlockHeight: fuzz.requestHeight,
		Reveal:             fuzz.reveal,
		GasUsed:            fuzz.gasUsed,
		ExitCode:           fuzz.exitCode,
		ProxyPubKeys:       []string{},
	}

	var revealMsgs [][]byte
	var commitments []string
	var revealProofs []string
	for i := 0; i < replicationFactor; i++ {
		revealMsg, commitment, revealProof, err := f.createRevealMsg(stakers[i], revealBody)
		require.NoError(t, err)

		commitProof, err := f.generateCommitProof(stakers[i].key, drID, commitment, res.Height)
		require.NoError(t, err)

		_, err = f.contractKeeper.Execute(
			f.Context(),
			f.coreContractAddr,
			stakers[i].address,
			commitMsg(drID, commitment, stakers[i].pubKey, commitProof, 0),
			sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
		)
		require.NoError(t, err)

		revealMsgs = append(revealMsgs, revealMsg)
		commitments = append(commitments, commitment)
		revealProofs = append(revealProofs, revealProof)
	}

	for i := 0; i < len(revealMsgs); i++ {
		msg := revealMsg(
			revealBody.RequestID,
			revealBody.Reveal,
			stakers[i].pubKey,
			revealProofs[i],
			revealBody.ProxyPubKeys,
			revealBody.ExitCode,
			revealBody.RequestBlockHeight,
			revealBody.GasUsed,
		)

		_, err = f.contractKeeper.Execute(
			f.Context(),
			f.coreContractAddr,
			stakers[i].address,
			msg,
			sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
		)
		require.NoError(t, err)
	}

	if timeout {
		for i := 0; i < defaultRevealTimeoutBlocks; i++ {
			f.AddBlock()
		}
	}

	return res.DrID, stakers
}
