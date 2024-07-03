package keeper

import (
	"encoding/base64"
	"encoding/json"
)

type Request struct {
	DrBinaryID        string                `json:"dr_binary_id"`
	DrInputs          string                `json:"dr_inputs"`
	GasLimit          string                `json:"gas_limit"`
	GasPrice          string                `json:"gas_price"`
	Height            uint64                `json:"height"`
	ID                string                `json:"id"`
	Memo              string                `json:"memo"`
	PaybackAddress    string                `json:"payback_address"`
	ReplicationFactor int64                 `json:"replication_factor"`
	ConsensusFilter   string                `json:"consensus_filter"`
	Reveals           map[string]RevealBody `json:"reveals"`
	SedaPayload       string                `json:"seda_payload"`
	TallyBinaryID     string                `json:"tally_binary_id"`
	TallyInputs       string                `json:"tally_inputs"`
	Version           string                `json:"version"`
}

type RevealBody struct {
	Salt     []byte `json:"salt"`
	ExitCode byte   `json:"exit_code"`
	GasUsed  string `json:"gas_used"`
	Reveal   string `json:"reveal"` // base64-encoded string
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

type VMResult struct {
	Salt        []byte `json:"salt"`
	ExitCode    byte   `json:"exit_code"`
	GasUsed     string `json:"gas_used"`
	Reveal      []byte `json:"reveal"`
	InConsensus byte   `json:"inConsensus"`
}

type Sudo struct {
	ID       string     `json:"dr_id"`
	Result   DataResult `json:"result"`
	ExitCode byte       `json:"exit_code"`
}

type DataResult struct {
	Version        string `json:"version"`
	ID             string `json:"dr_id"`
	BlockHeight    uint64 `json:"block_height"`
	ExitCode       byte   `json:"exit_code"`
	GasUsed        string `json:"gas_used"`
	Result         []byte `json:"result"`
	PaybackAddress string `json:"payback_address"`
	SedaPayload    string `json:"seda_payload"`
	Consensus      bool   `json:"consensus"`
	ModuleError    string `json:"module_error"` // error while processing filter or tally
}
