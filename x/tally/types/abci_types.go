package types

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm/v3"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
)

type ContractListResponse struct {
	IsPaused      bool          `json:"is_paused"`
	DataRequests  []Request     `json:"data_requests"`
	LastSeenIndex LastSeenIndex `json:"last_seen_index"`
}

// LastSeenIndex is a struct that contains the key of the last seen index in the data request
// list in the contract. The exact type of the key is not important so we simply store is as a
// string which can be used directly in the query.
type LastSeenIndex struct {
	index string
}

func (l *LastSeenIndex) UnmarshalJSON(data []byte) (err error) {
	if data == nil || string(data) == "null" {
		*l = LastSeenIndex{index: "null"}
	} else {
		*l = LastSeenIndex{index: string(data)}
	}

	return nil
}

func EmptyLastSeenIndex() LastSeenIndex {
	return LastSeenIndex{index: "null"}
}

func (l *LastSeenIndex) String() string {
	return l.index
}

type Request struct {
	ID                string                `json:"id"`
	Height            uint64                `json:"height"`
	ExecProgramID     string                `json:"exec_program_id"`
	ExecInputs        string                `json:"exec_inputs"`
	ExecGasLimit      uint64                `json:"exec_gas_limit"`
	TallyProgramID    string                `json:"tally_program_id"`
	TallyInputs       string                `json:"tally_inputs"`
	TallyGasLimit     uint64                `json:"tally_gas_limit"`
	PostedGasPrice    string                `json:"posted_gas_price"`
	Memo              string                `json:"memo"`
	PaybackAddress    string                `json:"payback_address"`
	ReplicationFactor uint16                `json:"replication_factor"`
	ConsensusFilter   string                `json:"consensus_filter"`
	Commits           map[string][]byte     `json:"commits"`
	Reveals           map[string]RevealBody `json:"reveals"`
	SedaPayload       string                `json:"seda_payload"`
	Version           string                `json:"version"`
}

// Validate validates the request fields and returns any validation error along
// with a partially filled DataResult containing the valid fields that were
// successfully decoded.
func (req *Request) ToResult(ctx types.Context) (result batchingtypes.DataResult, encodingErr error) {
	// If for whatever reason the request ID is not a valid hex string there is
	// no way to proceed, so we're simply going panic.
	// This should never happen since encoding bytes to hex is an operation that
	// can't fail on the contract (Rust) side.
	if _, err := hex.DecodeString(req.ID); err != nil {
		panic(fmt.Sprintf("invalid request ID: %s", req.ID))
	}

	result.DrId = req.ID
	result.DrBlockHeight = req.Height
	result.Version = req.Version
	//nolint:gosec // G115: We shouldn't get negative block heights.
	result.BlockHeight = uint64(ctx.BlockHeight())
	//nolint:gosec // G115: We shouldn't get negative timestamps.
	result.BlockTimestamp = uint64(ctx.BlockTime().Unix())

	// Validate PaybackAddress
	if _, err := base64.StdEncoding.DecodeString(req.PaybackAddress); err != nil {
		encodingErr = fmt.Errorf("invalid payback address: %w", err)
		result.PaybackAddress = base64.RawStdEncoding.EncodeToString([]byte(""))
	} else {
		result.PaybackAddress = req.PaybackAddress
	}

	// Validate SedaPayload
	if _, err := base64.StdEncoding.DecodeString(req.SedaPayload); err != nil {
		encodingErr = fmt.Errorf("invalid seda payload: %w", err)
		result.SedaPayload = base64.RawStdEncoding.EncodeToString([]byte(""))
	} else {
		result.SedaPayload = req.SedaPayload
	}

	return result, encodingErr
}

// SanitizeReveals returns sanitized reveals, executors, and gas reports. The
// three slices are sorted by executor with given entropies and have the same
// ordering. Each reveal's reported proxy public keys are also sorted.
func (req *Request) SanitizeReveals(height int64) ([]Reveal, []string, []uint64) {
	reveals := make([]Reveal, len(req.Reveals))
	i := 0
	for executor, reveal := range req.Reveals {
		reveals[i] = Reveal{Executor: executor, RevealBody: reveal}
		sort.Strings(reveals[i].ProxyPubKeys)
		i++
	}

	sortedReveals := HashSort(reveals, GetEntropy(req.ID, height))

	executors := make([]string, len(sortedReveals))
	gasReports := make([]uint64, len(sortedReveals))
	for i, reveal := range sortedReveals {
		executors[i] = reveal.Executor
		gasReports[i] = reveal.GasUsed
	}
	return sortedReveals, executors, gasReports
}

var _ HashSortable = Reveal{}

type Reveal struct {
	Executor string // executor ID (hex-encoded public key)
	RevealBody
}

func (r Reveal) GetSortKey() []byte {
	return []byte(r.Executor)
}

type RevealBody struct {
	DrID          string   `json:"dr_id"`
	DrBlockHeight uint64   `json:"dr_block_height"`
	ExitCode      byte     `json:"exit_code"`
	GasUsed       uint64   `json:"gas_used"`
	Reveal        string   `json:"reveal"` // base64-encoded string
	ProxyPubKeys  []string `json:"proxy_public_keys"`
}

func (u *RevealBody) MarshalJSON() ([]byte, error) {
	revealBytes, err := base64.StdEncoding.DecodeString(u.Reveal)
	if err != nil {
		return nil, err
	}

	type Alias RevealBody
	return json.Marshal(&struct {
		Reveal []int `json:"reveal"`
		*Alias
	}{
		Reveal: bytesToIntSlice(revealBytes),
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

func MarshalSudoRemoveDataRequests(processedReqs map[string][]Distribution) ([]byte, error) {
	return json.Marshal(struct {
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
}
