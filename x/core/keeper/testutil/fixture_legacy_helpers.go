package testutil

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
)

type Distribution struct {
	Burn            *DistributionBurn            `json:"burn,omitempty"`
	ExecutorReward  *DistributionExecutorReward  `json:"executor_reward,omitempty"`
	DataProxyReward *DistributionDataProxyReward `json:"data_proxy_reward,omitempty"`
}

type DistributionBurn struct {
	Amount math.Int `json:"amount"`
}

type DistributionDataProxyReward struct {
	PayoutAddress string   `json:"payout_address"`
	Amount        math.Int `json:"amount"`
	// The public key of the data proxy as a hex string
	PublicKey string `json:"public_key"`
}

type DistributionExecutorReward struct {
	Amount math.Int `json:"amount"`
	// The public key of the executor as a hex string
	Identity string `json:"identity"`
}

func (f *Fixture) SimulateLegacyTallyEndblock(t *testing.T, numBlocks int) {
	t.Helper()

	for i := 0; i < numBlocks; i++ {
		f.AddBlock()

		_, err := f.WasmKeeper.Sudo(f.Context(), f.CoreContractAddr, []byte(`{"expire_data_requests":{}}`))
		require.NoError(t, err)

		limit := 100
		queryRes, err := f.WasmViewKeeper.QuerySmart(
			f.Context(), f.CoreContractAddr,
			fmt.Appendf(nil, `{"get_data_requests_by_status":{"status": "tallying", "last_seen_index": %s, "limit": %d}}`, "null", limit),
		)
		require.NoError(t, err)

		type Request struct {
			ID string `json:"id"`
		}
		type ContractListResponse struct {
			DataRequests []Request `json:"data_requests"`
		}
		var contractQueryResponse ContractListResponse
		err = json.Unmarshal(queryRes, &contractQueryResponse)
		require.NoError(t, err)

		tallyList := contractQueryResponse.DataRequests
		processedReqs := make(map[string][]Distribution)
		for _, req := range tallyList {
			// Initialize the processedReqs map for each request with a full refund (no other distributions)
			processedReqs[req.ID] = make([]Distribution, 0)
		}

		// Notify the Core Contract of tally completion.
		removeDataRequestsMsg, err := json.Marshal(struct {
			SudoRemoveDataRequests struct {
				Requests map[string][]Distribution `json:"requests"`
			} `json:"remove_data_requests"`
		}{
			SudoRemoveDataRequests: struct {
				Requests map[string][]Distribution `json:"requests"`
			}{
				Requests: processedReqs,
			},
		})
		require.NoError(t, err)

		_, err = f.WasmKeeper.Sudo(f.Context(), f.CoreContractAddr, removeDataRequestsMsg)
		require.NoError(t, err)
	}
}
