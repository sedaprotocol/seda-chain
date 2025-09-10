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
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func TestEndBlock(t *testing.T) {
	f := testutil.InitFixture(t)

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
			err := f.SetDataProxyConfig(proxyPubKeys[0], "seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f", sdk.NewCoin(testutil.BondDenom, math.NewInt(1000000000000000000)))
			require.NoError(t, err)

			drID := f.ExecuteDataRequestFlow(
				t, nil, nil,
				tt.replicationFactor, tt.numCommits, tt.numReveals, tt.timeout,
				testutil.CommitRevealConfig{
					RequestHeight: 1,
					RequestMemo:   tt.memo,
					Reveal:        []byte("reveal"),
					ProxyPubKeys:  proxyPubKeys,
					GasUsed:       150000000000000000,
				})

			// Data request should be in the tallying status.
			dr, err := f.CoreKeeper.GetDataRequest(f.Context(), drID)
			require.NoError(t, err)
			require.Equal(t, types.DATA_REQUEST_STATUS_TALLYING, dr.Status)
			for executor := range dr.Reveals {
				_, err := f.CoreKeeper.GetRevealBody(f.Context(), drID, executor)
				require.NoError(t, err)
			}

			// TODO re-activate more detailed checks
			// beforeBalance := f.bankKeeper.GetBalance(f.Context(), f.stakers[0].address, bondDenom)
			// posterBeforeBalance := f.bankKeeper.GetBalance(f.Context(), f.deployer, bondDenom)

			err = f.CoreKeeper.EndBlock(f.Context())
			require.NoError(t, err)

			// Data request should have been removed from the store.
			dr, err = f.CoreKeeper.GetDataRequest(f.Context(), drID)
			require.Error(t, err)
			for executor := range dr.Reveals {
				_, err := f.CoreKeeper.GetRevealBody(f.Context(), drID, executor)
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
	f := testutil.InitFixture(t)

	execProgram := wasmstoragetypes.NewOracleProgram(testwasms.HTTPHeavyWasm(), f.Context().BlockTime())
	tallyProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm(), f.Context().BlockTime())

	tests := []struct {
		name               string
		CommitGasLimit     uint64
		RevealGasLimit     uint64
		commitExpectRefund bool
		revealExpectRefund bool
	}{
		{
			name:               "tx fees should be refunded",
			CommitGasLimit:     100000,
			RevealGasLimit:     100000,
			commitExpectRefund: true,
			revealExpectRefund: true,
		},
		{
			name:               "commit gas limit too large for refund",
			CommitGasLimit:     300000,
			RevealGasLimit:     100000,
			commitExpectRefund: false,
			revealExpectRefund: true,
		},
		{
			name:               "reveal gas limit too large for refund",
			CommitGasLimit:     150000,
			RevealGasLimit:     300000,
			commitExpectRefund: true,
			revealExpectRefund: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialBalance := f.BankKeeper.GetBalance(f.Context(), f.Stakers[0].Address, testutil.BondDenom)

			config := testutil.CommitRevealConfig{
				RequestHeight:  1,
				RequestMemo:    "",
				Reveal:         []byte("reveal"),
				ProxyPubKeys:   []string{},
				GasUsed:        150000000000000000,
				CommitGasLimit: tt.CommitGasLimit,
				RevealGasLimit: tt.RevealGasLimit,
			}

			res := f.PostDataRequest(
				t, execProgram.Hash, tallyProgram.Hash,
				base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", rand.Int63()))),
				1,
			)

			// Commit and check balance
			revealMsgs := f.CommitDataRequest(t, f.Stakers[:1], res.Height, res.DrID, config)

			afterCommitBalance := f.BankKeeper.GetBalance(f.Context(), f.Stakers[0].Address, testutil.BondDenom)
			diff := initialBalance.Sub(afterCommitBalance)
			if tt.commitExpectRefund {
				require.Equal(t, "0aseda", diff.String(), "tx fee must have been refunded")
			} else {
				fee := sdk.NewCoins(sdk.NewCoin(testutil.BondDenom, math.NewIntFromUint64(tt.CommitGasLimit).Mul(math.NewInt(1e10))))
				require.Equal(t, fee.String(), diff.String(), "tx fee must have been deducted")
			}

			// Reveal and check balance
			f.ExecuteReveals(t, f.Stakers[:1], revealMsgs, config)

			afterRevealBalance := f.BankKeeper.GetBalance(f.Context(), f.Stakers[0].Address, testutil.BondDenom)
			diff = afterCommitBalance.Sub(afterRevealBalance)
			if tt.revealExpectRefund {
				require.Equal(t, "0aseda", diff.String(), "tx fee must have been refunded")
			} else {
				fee := sdk.NewCoins(sdk.NewCoin(testutil.BondDenom, math.NewIntFromUint64(tt.RevealGasLimit).Mul(math.NewInt(1e10))))
				require.Equal(t, fee.String(), diff.String(), "tx fee must have been deducted")
			}
		})
	}
}
