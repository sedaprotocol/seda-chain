package keeper_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func TestEndBlock(t *testing.T) {
	f := initFixture(t)

	tests := []struct {
		name              string
		memo              string
		replicationFactor int
		numCommits        int
		numReveals        int
		timeout           bool
		expExitCode       uint32
	}{
		{
			name:              "full single commit-reveal",
			memo:              base64.StdEncoding.EncodeToString([]byte("memo0")),
			replicationFactor: 1,
			numCommits:        1,
			numReveals:        1,
			timeout:           false,
			expExitCode:       0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxyPubKeys := []string{"03b27f2df0cbdb5cdadff5b4be0c9fda5aa3a59557ef6d0b49b4298ef42c8ce2b0"}
			err := f.SetDataProxyConfig(proxyPubKeys[0], "seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f", sdk.NewCoin(bondDenom, math.NewInt(1000000000000000000)))
			require.NoError(t, err)

			drID := f.executeDataRequestFlow(
				t, nil, nil,
				tt.replicationFactor, tt.numCommits, tt.numReveals, tt.timeout,
				commitRevealConfig{
					requestHeight: 1,
					requestMemo:   tt.memo,
					reveal:        []byte("reveal"),
					proxyPubKeys:  proxyPubKeys,
					gasUsed:       150000000000000000,
				})

			// Data request should be in the tallying status.
			dr, err := f.coreKeeper.GetDataRequest(f.Context(), drID)
			require.NoError(t, err)
			require.Equal(t, types.DATA_REQUEST_STATUS_TALLYING, dr.Status)
			for executor := range dr.Reveals {
				_, err := f.coreKeeper.GetRevealBody(f.Context(), drID, executor)
				require.NoError(t, err)
			}

			// TODO re-activate more detailed checks
			// beforeBalance := f.bankKeeper.GetBalance(f.Context(), f.stakers[0].address, bondDenom)
			// posterBeforeBalance := f.bankKeeper.GetBalance(f.Context(), f.deployer, bondDenom)

			err = f.coreKeeper.EndBlock(f.Context())
			require.NoError(t, err)

			// Data request should have been removed from the store.
			dr, err = f.coreKeeper.GetDataRequest(f.Context(), drID)
			require.Error(t, err)
			for executor := range dr.Reveals {
				_, err := f.coreKeeper.GetRevealBody(f.Context(), drID, executor)
				require.Error(t, err)
			}

			// // TODO query get_staker pending_withdrawal and check diff
			// // Verify the staker did not pay for the transactions
			// afterBalance := f.bankKeeper.GetBalance(f.Context(), f.stakers[0].address, bondDenom)
			// diff := afterBalance.Sub(beforeBalance)
			// require.Equal(t, "0aseda", diff.String())

			// // Verify the poster paid for execution
			// afterPostBalance := f.bankKeeper.GetBalance(f.Context(), f.deployer, bondDenom)
			// diff = afterPostBalance.Sub(posterBeforeBalance)
			// require.NotEqual(t, "0aseda", diff.String(), "Poster should have paid for execution")

			// dataResult, err := f.batchingKeeper.GetLatestDataResult(f.Context(), drID)
			// require.NoError(t, err)
			// // TODO map oracle program to exit code
			// // require.Equal(t, tt.expExitCode, dataResult.ExitCode)

			// dataResults, err := f.batchingKeeper.GetDataResults(f.Context(), false)
			// require.NoError(t, err)
			// require.Contains(t, dataResults, *dataResult)
		})
	}
}

func TestTxFeeRefund(t *testing.T) {
	f := initFixture(t)

	execProgram := wasmstoragetypes.NewOracleProgram(testwasms.HTTPHeavyWasm(), f.Context().BlockTime())
	tallyProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm(), f.Context().BlockTime())

	tests := []struct {
		name               string
		commitGasLimit     uint64
		revealGasLimit     uint64
		commitExpectRefund bool
		revealExpectRefund bool
	}{
		{
			name:               "tx fees should be refunded",
			commitGasLimit:     100000,
			revealGasLimit:     100000,
			commitExpectRefund: true,
			revealExpectRefund: true,
		},
		{
			name:               "commit gas limit too large for refund",
			commitGasLimit:     300000,
			revealGasLimit:     100000,
			commitExpectRefund: false,
			revealExpectRefund: true,
		},
		{
			name:               "reveal gas limit too large for refund",
			commitGasLimit:     150000,
			revealGasLimit:     300000,
			commitExpectRefund: true,
			revealExpectRefund: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialBalance := f.bankKeeper.GetBalance(f.Context(), f.stakers[0].address, bondDenom)

			config := commitRevealConfig{
				requestHeight:  1,
				requestMemo:    "",
				reveal:         []byte("reveal"),
				proxyPubKeys:   []string{},
				gasUsed:        150000000000000000,
				commitGasLimit: tt.commitGasLimit,
				revealGasLimit: tt.revealGasLimit,
			}

			res, err := f.postDataRequest(
				t, execProgram.Hash, tallyProgram.Hash,
				base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", rand.Int63()))),
				1,
			)
			require.NoError(t, err)

			// Commit and check balance
			revealMsgs, err := f.commitDataRequest(t, f.stakers[:1], res.Height, res.DrID, config)
			require.NoError(t, err)

			afterCommitBalance := f.bankKeeper.GetBalance(f.Context(), f.stakers[0].address, bondDenom)
			diff := initialBalance.Sub(afterCommitBalance)
			if tt.commitExpectRefund {
				require.Equal(t, "0aseda", diff.String(), "tx fee must have been refunded")
			} else {
				fee := sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(tt.commitGasLimit).Mul(math.NewInt(1e10))))
				require.Equal(t, fee.String(), diff.String(), "tx fee must have been deducted")
			}

			// Reveal and check balance
			err = f.executeReveals(t, f.stakers[:1], revealMsgs, config)
			require.NoError(t, err)

			afterRevealBalance := f.bankKeeper.GetBalance(f.Context(), f.stakers[0].address, bondDenom)
			diff = afterCommitBalance.Sub(afterRevealBalance)
			if tt.revealExpectRefund {
				require.Equal(t, "0aseda", diff.String(), "tx fee must have been refunded")
			} else {
				fee := sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(tt.revealGasLimit).Mul(math.NewInt(1e10))))
				require.Equal(t, fee.String(), diff.String(), "tx fee must have been deducted")
			}
		})
	}
}
