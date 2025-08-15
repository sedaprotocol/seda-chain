package types

import (
	"encoding/json"

	"cosmossdk.io/math"

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm/v2"
)

var _ HashSortable = Reveal{}

type Reveal struct {
	Executor string // executor ID (hex-encoded public key)
	RevealBody
}

func (r Reveal) GetSortKey() []byte {
	return []byte(r.Executor)
}

func (u *RevealBody) MarshalJSON() ([]byte, error) {
	type Alias RevealBody
	return json.Marshal(&struct {
		Reveal []int `json:"reveal"`
		*Alias
	}{
		Reveal: bytesToIntSlice(u.Reveal),
		Alias:  (*Alias)(u),
	})
}

func bytesToIntSlice(bytes []byte) []int {
	intSlice := make([]int, len(bytes))
	for i, b := range bytes {
		intSlice[i] = int(b)
	}
	return intSlice
}

type VMResult struct {
	Stdout      []string
	Stderr      []string
	Result      []byte
	GasUsed     uint64
	ExitCode    uint32
	ExitMessage string
}

// MapVMResult maps a tallyvm.VmResult to a VmResult, taking care of checking the result pointer
// and setting the exit message if the result is empty. This allows us to display the exit message
// to the end user.
func MapVMResult(vmRes tallyvm.VmResult) VMResult {
	result := VMResult{
		//nolint:gosec // G115: We shouldn't get negative exit code anyway.
		ExitCode:    uint32(vmRes.ExitInfo.ExitCode),
		ExitMessage: vmRes.ExitInfo.ExitMessage,
		Stdout:      vmRes.Stdout,
		Stderr:      vmRes.Stderr,
		GasUsed:     vmRes.GasUsed,
	}

	if vmRes.Result == nil || (vmRes.ResultLen == 0 && vmRes.ExitInfo.ExitCode != 0) {
		result.Result = []byte(vmRes.ExitInfo.ExitMessage)
	} else if vmRes.Result != nil {
		result.Result = *vmRes.Result
	}

	return result
}

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

func NewBurn(amount, gasPrice math.Int) Distribution {
	return Distribution{
		Burn: &DistributionBurn{Amount: amount.Mul(gasPrice)},
	}
}

func NewDataProxyReward(pubkey, payoutAddr string, amount, gasPrice math.Int) Distribution {
	return Distribution{
		DataProxyReward: &DistributionDataProxyReward{
			PayoutAddress: payoutAddr,
			Amount:        amount.Mul(gasPrice),
			PublicKey:     pubkey,
		},
	}
}

func NewExecutorReward(identity string, amount, gasPrice math.Int) Distribution {
	return Distribution{
		ExecutorReward: &DistributionExecutorReward{
			Identity: identity,
			Amount:   amount.Mul(gasPrice),
		},
	}
}
