package types

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"

	"golang.org/x/crypto/sha3"

	"cosmossdk.io/math"
)

type Request struct {
	ID                string                `json:"id"`
	Height            uint64                `json:"height"`
	ExecProgramID     string                `json:"exec_program_id"`
	ExecInputs        string                `json:"exec_inputs"`
	ExecGasLimit      uint64                `json:"exec_gas_limit"`
	TallyProgramID    string                `json:"tally_program_id"`
	TallyInputs       string                `json:"tally_inputs"`
	TallyGasLimit     uint64                `json:"tally_gas_limit"`
	GasPrice          string                `json:"gas_price"`
	Memo              string                `json:"memo"`
	PaybackAddress    string                `json:"payback_address"`
	ReplicationFactor uint16                `json:"replication_factor"`
	ConsensusFilter   string                `json:"consensus_filter"`
	Commits           map[string][]byte     `json:"commits"`
	Reveals           map[string]RevealBody `json:"reveals"`
	SedaPayload       string                `json:"seda_payload"`
	Version           string                `json:"version"`
}

type RevealBody struct {
	ID           string   `json:"id"`
	Salt         []byte   `json:"salt"`
	ExitCode     byte     `json:"exit_code"`
	GasUsed      uint64   `json:"gas_used"`
	Reveal       string   `json:"reveal"` // base64-encoded string
	ProxyPubKeys []string `json:"proxy_public_keys"`
}

func (u *RevealBody) MarshalJSON() ([]byte, error) {
	revealBytes, err := base64.StdEncoding.DecodeString(u.Reveal)
	if err != nil {
		return nil, err
	}

	intSlice := make([]int, len(revealBytes))
	for i, b := range revealBytes {
		intSlice[i] = int(b)
	}

	saltIntSlice := make([]int, len(u.Salt))
	for i, b := range u.Salt {
		saltIntSlice[i] = int(b)
	}

	type Alias RevealBody
	return json.Marshal(&struct {
		Reveal []int `json:"reveal"`
		Salt   []int `json:"salt"`
		*Alias
	}{
		Reveal: intSlice,
		Salt:   saltIntSlice,
		Alias:  (*Alias)(u),
	})
}

func (u *RevealBody) TryHash() (string, error) {
	revealHasher := sha3.NewLegacyKeccak256()
	revealBytes, err := base64.StdEncoding.DecodeString(u.Reveal)
	if err != nil {
		return "", err
	}
	revealHasher.Write(revealBytes)
	revealHash := revealHasher.Sum(nil)

	hasher := sha3.NewLegacyKeccak256()

	idBytes, err := hex.DecodeString(u.ID)
	if err != nil {
		return "", err
	}
	hasher.Write(idBytes)

	hasher.Write(u.Salt)
	hasher.Write([]byte{u.ExitCode})

	gasUsedBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(gasUsedBytes, u.GasUsed)
	hasher.Write(gasUsedBytes)

	hasher.Write(revealHash)

	proxyPubKeyHasher := sha3.NewLegacyKeccak256()
	for _, key := range u.ProxyPubKeys {
		keyHasher := sha3.NewLegacyKeccak256()
		keyHasher.Write([]byte(key))
		proxyPubKeyHasher.Write(keyHasher.Sum(nil))
	}
	hasher.Write(proxyPubKeyHasher.Sum(nil))

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

type PayoutRecord struct {
	Burn       math.Int       `json:"burn"`
	ProxyDists []Distribution `json:"proxy_dists"`
	ExecDists  []Distribution `json:"exec_dists"`
}

// DistributionsWithBaseFee constructs and returns a slice of distributions that
// the contract expects based on the payout record struct and a new base fee burn.
func (p PayoutRecord) DistributionsWithBaseFee(baseFee math.Int) []Distribution {
	if p.Burn.IsNil() {
		p.Burn = math.ZeroInt()
	}
	p.Burn = p.Burn.Add(baseFee)
	dists := []Distribution{}
	dists = append(dists, NewBurn(p.Burn))
	dists = append(dists, p.ProxyDists...)
	dists = append(dists, p.ExecDists...)
	return dists
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
	To     []byte   `json:"to"`
	Amount math.Int `json:"amount"`
}

type DistributionExecutorReward struct {
	Amount   math.Int `json:"amount"`
	Identity string   `json:"identity"`
}

func NewBurn(amount math.Int) Distribution {
	return Distribution{
		Burn: &DistributionBurn{Amount: amount},
	}
}

func NewDataProxyReward(pubKey []byte, amount math.Int) Distribution {
	return Distribution{
		DataProxyReward: &DistributionDataProxyReward{
			To:     pubKey,
			Amount: amount,
		},
	}
}

func NewExecutorReward(identity string, amount math.Int) Distribution {
	return Distribution{
		ExecutorReward: &DistributionExecutorReward{
			Identity: identity,
			Amount:   amount,
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
